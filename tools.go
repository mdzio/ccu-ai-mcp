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
			resp, err := scriptClient.ExecuteTmpl(templ, params)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("Execution of the HM script for the tool failed: %v", err),
						},
					},
					IsError: true,
				}, nil
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
						Type: "text",
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
