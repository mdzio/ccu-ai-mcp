package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mdzio/ccu-ai-mcp/config"
	"github.com/mdzio/go-hmccu/v2/script"
)

const (
	// working directory for shell scripts on the CCU
	shellScriptDir = "/tmp"
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

	var serverTools []server.ServerTool
	for _, toolCfg := range configTools.Tools {
		// only enabled tools
		if !toolCfg.Enabled {
			continue
		}

		toolsLog.Debug("Creating tool", "name", toolCfg.Name, "kind", toolCfg.Kind)
		if toolCfg.Kind == config.ShellScript && !configMain.CCU.OnCCU {
			toolsLog.Warn("Skipping shell script tool, MCP server runs not on the CCU", "name", toolCfg.Name)
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
				return handleHMScript(templ, scriptClient, request)
			}
		case config.ShellScript:
			handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return handleShellScript(templ, request)
			}
		default:
			return nil, fmt.Errorf("unsupported tool kind %s in tool %s", toolCfg.Kind, toolCfg.Name)
		}

		// wrap the handler, so that every tool call is logged with its
		// name and parameters
		toolName := toolCfg.Name
		toolHandler := handler
		handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			toolsLog.Debug("MCP tool call received", append([]any{"name", toolName}, mapAsSlice(request.GetArguments())...)...)
			return toolHandler(ctx, request)
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
func handleHMScript(templ *template.Template, scriptClient *script.Client, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// fill the script template with the parameters of the MCP call
	scriptText, err := fillTemplate(templ, request.GetArguments())
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
func handleShellScript(templ *template.Template, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// fill the script template with the parameters of the MCP call
	scriptText, err := fillTemplate(templ, request.GetArguments())
	if err != nil {
		return textResult("ERROR: Rendering of the shell script template failed: "+err.Error(), true), nil
	}

	// execute the script in the working directory /tmp, stdout and stderr
	// are captured into the same buffer to preserve their order
	toolsLog.Debug("Executing shell script", "script", scriptText)
	var output bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", scriptText)
	cmd.Dir = shellScriptDir
	cmd.Stdout = &output
	cmd.Stderr = &output
	err = cmd.Run()

	// build result text, an error description is prepended if the execution
	// failed
	var sb strings.Builder
	var exitErr *exec.ExitError
	switch {
	case err == nil:
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
	// exit code of the script, -1 if it could not be determined
	exitCode := 0
	if exitErr != nil {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		exitCode = -1
	}
	toolsLog.Debug("Shell script response", "response", sb.String(), "exitCode", exitCode)
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

// textResult creates a tool result with a single text content.
func textResult(text string, isError bool) *mcp.CallToolResult {
	if isError {
		return mcp.NewToolResultError(text)
	}
	return mcp.NewToolResultText(text)
}

// mapAsSlice converts a map[string]any to []any for slog
func mapAsSlice(m map[string]any) []any {
	attrs := make([]any, 0, len(m)*2)
	for k, v := range m {
		attrs = append(attrs, k, v)
	}
	return attrs
}
