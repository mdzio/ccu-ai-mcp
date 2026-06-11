package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mdzio/ccu-ai-mcp/ccu"
	"github.com/mdzio/ccu-ai-mcp/config"
)

// createTools creates all MCP tools for the server
func createTools() ([]server.ServerTool, error) {
	// create script client
	scriptClient := &ccu.ScriptClient{
		Addr:     configMain.CCU.Address,
		UserName: configMain.CCU.User,
		Password: configMain.CCU.Password,
	}

	var serverTools []server.ServerTool
	for _, toolCfg := range configTools.Tools {
		// only enabled tools
		if !toolCfg.Enabled {
			continue
		}

		// only HM scripts are currently supported
		slog.Debug("Creating tool", "name", toolCfg.Name, "kind", toolCfg.Kind)
		if toolCfg.Kind != config.HMScript {
			return nil, fmt.Errorf("unsupported tool kind %s in tool %s", toolCfg.Kind, toolCfg.Name)
		}

		// create tool options
		opts := []mcp.ToolOption{
			mcp.WithDescription(toolCfg.Description),
		}

		// add parameters
		for _, param := range toolCfg.Parameters {
			switch param.Type {
			case config.String:
				opts = append(opts, mcp.WithString(param.Name,
					mcp.Description(param.Description),
					mcp.Required()))
			case config.Number:
				opts = append(opts, mcp.WithNumber(param.Name,
					mcp.Description(param.Description),
					mcp.Required()))
			case config.Integer:
				opts = append(opts, mcp.WithInteger(param.Name,
					mcp.Description(param.Description),
					mcp.Required()))
			case config.Boolean:
				opts = append(opts, mcp.WithBoolean(param.Name,
					mcp.Description(param.Description),
					mcp.Required()))
			case config.StringArray:
				opts = append(opts, mcp.WithArray(param.Name,
					mcp.WithStringItems(),
					mcp.Description(param.Description),
					mcp.Required()))
			case config.IntegerArray:
				opts = append(opts, mcp.WithArray(param.Name,
					mcp.WithIntegerItems(),
					mcp.Description(param.Description),
					mcp.Required()))
			case config.NumberArray:
				opts = append(opts, mcp.WithArray(param.Name,
					mcp.WithNumberItems(),
					mcp.Description(param.Description),
					mcp.Required()))
			case config.BooleanArray:
				opts = append(opts, mcp.WithArray(param.Name,
					mcp.WithBooleanItems(),
					mcp.Description(param.Description),
					mcp.Required()))
			case config.Any:
				opts = append(opts, mcp.WithAny(param.Name,
					mcp.Description(param.Description),
					mcp.Required()))
			default:
				return nil, fmt.Errorf("unsupported parameter type %s in tool %s", param.Type, toolCfg.Name)
			}
		}

		// parse script as template
		templ, err := template.New(toolCfg.Name).Parse(toolCfg.Script)
		if err != nil {
			return nil, fmt.Errorf("parsing HM script template failed for tool %s: %w", toolCfg.Name, err)
		}

		// create the tool
		mcpTool := mcp.NewTool(toolCfg.Name, opts...)

		// create handler for this tool
		handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			slog.Debug("Executing tool", "name", toolCfg.Name)

			// extract parameters from MCP call
			params := request.GetArguments()

			// execute the script template with parameters
			slog.Debug("Executing script with parameters", mapAsSlice(params)...)
			resp, err := scriptClient.ExecuteTmpl(templ, params)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Execution of the HM script for the tool failed", err), nil
			}

			// check for empty response
			if len(resp) == 0 || (len(resp) == 1 && resp[0] == "") {
				return mcp.NewToolResultError("Execution of the HM script returned empty result, the user should check the HM script of this tool"), nil
			}

			// build result text
			var resultText strings.Builder
			isError := false
			for i, line := range resp {
				if strings.HasPrefix(line, "ERROR:") {
					isError = true
				}
				if i > 0 {
					resultText.WriteString("\n")
				}
				resultText.WriteString(line)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: mcp.ContentTypeText,
						Text: resultText.String(),
					},
				},
				IsError: isError,
			}, nil
		}

		// add to server tools slice
		serverTools = append(serverTools, server.ServerTool{
			Tool:    mcpTool,
			Handler: handler,
		})
	}
	return serverTools, nil
}

// mapAsSlice converts a map[string]any to []any for slog
func mapAsSlice(m map[string]any) []any {
	attrs := make([]any, 0, len(m)*2)
	for k, v := range m {
		attrs = append(attrs, k, v)
	}
	return attrs
}
