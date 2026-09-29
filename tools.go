package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mdzio/ccu-ai-mcp/config"
	"github.com/mdzio/go-hmccu/v2/script"
)

const (
	// working directory for shell scripts on the CCU
	shellScriptDir = "/tmp"
	// maximum execution time of a shell script
	shellScriptTimeout = 60 * time.Second
	// maximum size of the captured output of a shell script
	shellOutputLimit = 1024 * 1024
	// grace period after killing a shell script for closing the output pipes
	shellWaitDelay = 5 * time.Second
)

var toolsLog = slog.With("component", "tools")

// shellFuncs are additional template functions for shell script templates.
var shellFuncs = template.FuncMap{
	"shq": shellQuote,
}

// createTools creates all MCP tools for the server
func createTools() ([]server.ServerTool, error) {
	// create script client
	scriptClient := &script.Client{
		Addr:               configMain.CCU.Address,
		UserName:           configMain.CCU.User,
		Password:           configMain.CCU.Password,
		UseTLS:             configMain.CCU.UseTLS,
		InsecureSkipVerify: configMain.CCU.InsecureSkipVerify,
		UseInternalPort:    configMain.CCU.UseInternalPort,
	}

	// shell scripts can only be executed, if the server runs on the CCU itself
	// (e.g. installed as add-on)
	localCCU := isLocalAddress(configMain.CCU.Address)

	var serverTools []server.ServerTool
	for _, toolCfg := range configTools.Tools {
		// only enabled tools
		if !toolCfg.Enabled {
			continue
		}

		toolsLog.Debug("Creating tool", "name", toolCfg.Name, "kind", toolCfg.Kind)
		if toolCfg.Kind == config.ShellScript && !localCCU {
			toolsLog.Warn("Skipping shell script tool, shell scripts can only be executed on the CCU itself", "name", toolCfg.Name)
			continue
		}

		// create tool options
		opts := []mcp.ToolOption{
			mcp.WithDescription(toolCfg.Description),
		}

		// add parameters
		for _, param := range toolCfg.Parameters {
			var propertyOpts []mcp.PropertyOption
			propertyOpts = append(propertyOpts, mcp.Description(param.Description))
			if !param.Optional {
				propertyOpts = append(propertyOpts, mcp.Required())
			}
			switch param.Type {
			case config.String:
				opts = append(opts, mcp.WithString(param.Name, propertyOpts...))
			case config.Number:
				opts = append(opts, mcp.WithNumber(param.Name, propertyOpts...))
			case config.Integer:
				opts = append(opts, mcp.WithInteger(param.Name, propertyOpts...))
			case config.Boolean:
				opts = append(opts, mcp.WithBoolean(param.Name, propertyOpts...))
			case config.StringArray:
				opts = append(opts, mcp.WithArray(param.Name, append([]mcp.PropertyOption{mcp.WithStringItems()}, propertyOpts...)...))
			case config.IntegerArray:
				opts = append(opts, mcp.WithArray(param.Name, append([]mcp.PropertyOption{mcp.WithIntegerItems()}, propertyOpts...)...))
			case config.NumberArray:
				opts = append(opts, mcp.WithArray(param.Name, append([]mcp.PropertyOption{mcp.WithNumberItems()}, propertyOpts...)...))
			case config.BooleanArray:
				opts = append(opts, mcp.WithArray(param.Name, append([]mcp.PropertyOption{mcp.WithBooleanItems()}, propertyOpts...)...))
			case config.Any:
				opts = append(opts, mcp.WithAny(param.Name, propertyOpts...))
			default:
				return nil, fmt.Errorf("unsupported parameter type %s in tool %s", param.Type, toolCfg.Name)
			}
		}

		// parse script as template, shell scripts get additional template
		// functions (e.g. shq)
		templ := template.New(toolCfg.Name)
		if toolCfg.Kind == config.ShellScript {
			templ = templ.Funcs(shellFuncs)
		}
		if _, err := templ.Parse(toolCfg.Script); err != nil {
			return nil, fmt.Errorf("parsing script template failed for tool %s: %w", toolCfg.Name, err)
		}

		// create the tool
		mcpTool := mcp.NewTool(toolCfg.Name, opts...)

		// create handler for this tool depending on its kind
		var handler server.ToolHandlerFunc
		switch toolCfg.Kind {
		case config.HMScript:
			handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return handleHMScript(toolCfg, templ, scriptClient, request)
			}
		case config.ShellScript:
			handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return handleShellScript(ctx, toolCfg, templ, request)
			}
		default:
			return nil, fmt.Errorf("unsupported tool kind %s in tool %s", toolCfg.Kind, toolCfg.Name)
		}

		// add to server tools slice
		serverTools = append(serverTools, server.ServerTool{
			Tool:    mcpTool,
			Handler: handler,
		})
	}
	return serverTools, nil
}

// handleHMScript executes a HM script tool on the CCU.
func handleHMScript(toolCfg config.Tool, templ *template.Template, scriptClient *script.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolsLog.Debug("Executing HM script tool", "name", toolCfg.Name)

	// fill the script template with the parameters of the MCP call
	params := request.GetArguments()
	toolsLog.Debug("Filling HM script template with parameters", mapAsSlice(params)...)
	scriptText, err := fillTemplate(templ, params)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Rendering of the HM script template failed", err), nil
	}

	// execute the script on the CCU
	resp, err := scriptClient.Execute(scriptText)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Execution of the HM script for the tool failed", err), nil
	}

	// check for empty response
	if len(resp) == 0 || (len(resp) == 1 && resp[0] == "") {
		return mcp.NewToolResultError("Execution of the HM script returned empty result, the HM script may have a syntax error or a called method does not exist"), nil
	}

	// a line starting with ERROR: signals an error
	isError := false
	for _, line := range resp {
		if strings.HasPrefix(line, "ERROR:") {
			isError = true
			break
		}
	}
	return textResult(strings.Join(resp, "\n"), isError), nil
}

// handleShellScript executes a shell script tool locally on the CCU.
func handleShellScript(ctx context.Context, toolCfg config.Tool, templ *template.Template, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolsLog.Debug("Executing shell script tool", "name", toolCfg.Name)

	// fill the script template with the parameters of the MCP call
	params := request.GetArguments()
	toolsLog.Debug("Filling shell script template with parameters", mapAsSlice(params)...)
	scriptText, err := fillTemplate(templ, params)
	if err != nil {
		return textResult("ERROR: Rendering of the shell script template failed: "+err.Error(), true), nil
	}

	// execute the script with a timeout in the working directory /tmp, stdout
	// and stderr are captured into the same buffer to preserve their order
	ctx, cancel := context.WithTimeout(ctx, shellScriptTimeout)
	defer cancel()
	output := &limitedBuffer{limit: shellOutputLimit}
	cmd := exec.CommandContext(ctx, "sh", "-c", scriptText)
	cmd.Dir = shellScriptDir
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.WaitDelay = shellWaitDelay
	err = cmd.Run()

	// build result text, an error description is prepended if the execution
	// failed
	var sb strings.Builder
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		fmt.Fprintf(&sb, "ERROR: Shell script timed out after %v\n", shellScriptTimeout)
	case ctx.Err() != nil:
		sb.WriteString("ERROR: Shell script execution was cancelled\n")
	case errors.As(err, &exitErr):
		fmt.Fprintf(&sb, "ERROR: Shell script failed with exit code %d\n", exitErr.ExitCode())
	default:
		fmt.Fprintf(&sb, "ERROR: Shell script execution failed: %v\n", err)
	}
	if output.Len() == 0 && err == nil {
		sb.WriteString("Shell script finished without output.")
	} else {
		sb.Write(output.Bytes())
	}
	if output.truncated {
		fmt.Fprintf(&sb, "\n[Output truncated after %d bytes]", shellOutputLimit)
	}
	return textResult(sb.String(), err != nil), nil
}

// fillTemplate executes the template with the specified data and returns the
// result.
func fillTemplate(templ *template.Template, data map[string]any) (string, error) {
	var sb strings.Builder
	if err := templ.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("rendering of template with data %v failed: %w", data, err)
	}
	return sb.String(), nil
}

// shellQuote quotes a value for safe use as a single word in a POSIX shell
// script. Slices are quoted element by element and separated by spaces. A nil
// value results in an empty word.
func shellQuote(v any) string {
	if list, ok := v.([]any); ok {
		words := make([]string, len(list))
		for i, e := range list {
			words[i] = shellQuote(e)
		}
		return strings.Join(words, " ")
	}
	var s string
	switch v := v.(type) {
	case nil:
	case float64:
		// JSON numbers are decoded as float64, avoid exponent notation
		s = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		s = fmt.Sprint(v)
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// isLocalAddress returns true, if the address (optionally with port) refers to
// the local host.
func isLocalAddress(addr string) bool {
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return strings.EqualFold(host, "localhost")
}

// textResult creates a tool result with a single text content.
func textResult(text string, isError bool) *mcp.CallToolResult {
	if isError {
		return mcp.NewToolResultError(text)
	}
	return mcp.NewToolResultText(text)
}

// limitedBuffer collects written data up to a limit and silently discards the
// rest.
type limitedBuffer struct {
	bytes.Buffer
	limit     int
	truncated bool
}

// Write implements io.Writer.
func (b *limitedBuffer) Write(p []byte) (int, error) {
	if room := b.limit - b.Len(); len(p) > room {
		b.Buffer.Write(p[:room])
		b.truncated = true
	} else {
		b.Buffer.Write(p)
	}
	return len(p), nil
}

// mapAsSlice converts a map[string]any to []any for slog
func mapAsSlice(m map[string]any) []any {
	attrs := make([]any, 0, len(m)*2)
	for k, v := range m {
		attrs = append(attrs, k, v)
	}
	return attrs
}
