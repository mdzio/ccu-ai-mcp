package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var httpLog = slog.With("component", "http")

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// authTokenKey is the context key for the authentication token
const authTokenKey contextKey = "authToken"

// authHTTPContextFunc extracts the Bearer token from the Authorization header
// and stores it in the request context
func authHTTPContextFunc(ctx context.Context, r *http.Request) context.Context {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx
	}

	// check if it's a Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ctx
	}

	// extract the token value
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return context.WithValue(ctx, authTokenKey, token)
}

// authMiddleware creates a middleware that validates the Bearer token from context
func authMiddleware(apiKey string) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// skip authentication if apiKey is empty
			if apiKey == "" {
				return next(ctx, request)
			}

			// get token from context
			token, ok := ctx.Value(authTokenKey).(string)
			if !ok || token == "" {
				return mcp.NewToolResultError("Authentication required: Bearer token missing"), nil
			}

			// validate token against configured apiKey
			if token != apiKey {
				return mcp.NewToolResultError("Authentication failed: Invalid API key"), nil
			}

			return next(ctx, request)
		}
	}
}

// serveHTTP serves MCP requests over HTTP or HTTPS
func serveHTTP(mcpServer *server.MCPServer, useTLS bool) error {
	// create streamable HTTP server with CORS configuration
	opts := []server.StreamableHTTPOption{
		server.WithStreamableHTTPCORS(
			server.WithCORSAllowedOrigins(configMain.MCP.CORSAllowedOrigins...),
			server.WithCORSAllowCredentials(),
		),
		server.WithHTTPContextFunc(authHTTPContextFunc),
	}

	// add TLS configuration if HTTPS transport is selected
	if useTLS {
		opts = append(opts, server.WithTLSCert(configMain.MCP.CertFile, configMain.MCP.KeyFile))
	} else {
		httpLog.Warn("No TLS (HTTPS) configured for MCP server")
	}

	// add authentication middleware if apiKey is configured
	if configMain.MCP.APIKey != "" {
		mcpServer.Use(authMiddleware(configMain.MCP.APIKey))
	} else {
		httpLog.Warn("No API key configured for MCP server")
	}

	// create and configure HTTP server
	httpServer := server.NewStreamableHTTPServer(mcpServer, opts...)

	// configure HTTP server address
	addr := fmt.Sprintf(":%d", configMain.MCP.Port)

	// react on INT or TERM signal
	signal.Notify(termSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		// wait for start up to complete
		time.Sleep(1 * time.Second)
		// wait for signal
		<-termSig
		// shutdown server
		httpServer.Shutdown(context.Background())
	}()

	// start serving
	httpLog.Info("Starting MCP HTTP(S) server", "address", addr, "endpoint", "/mcp")
	return httpServer.Start(addr)
}
