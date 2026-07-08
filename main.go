package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/mdzio/ccu-ai-mcp/config"
)

const (
	appName        = "ccu-ai-mcp"
	appDisplayName = "CCU-AI-MCP"
	appCopyright   = "©2026"
	appVendor      = "info@ccu-historian.de"
	appDescription = "CCU-AI-MCP is an MCP server for the OpenCCU smart home hub (https://openccu.de)."
)

var (
	appVersion  = "-dev-" // overwritten during build process
	configPath  = flag.String("config", "config.toml", "path to configuration `file`")
	configMain  *config.Main
	configTools *config.Tools
	// to ensure that no signal is missed, the buffer size must be 1
	termSig = make(chan os.Signal, 1)
	mainLog = slog.With("component", "main")
)

// configure handles all application configuration
func configure() error {
	// parse command line
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage of "+appName+":")
		flag.PrintDefaults()
	}
	// flag.Parse calls os.Exit(2) on error
	flag.Parse()

	// load configuration
	var err error
	configMain, err = config.Load(*configPath)
	if err != nil {
		return err
	}

	// set log level from config
	slog.SetLogLoggerLevel(configMain.General.LogLevel)

	// log startup message
	mainLog.Info("Starting " + appDisplayName + " V" + appVersion + " " + appCopyright + " " + appVendor)

	// load tools configuration
	configTools, err = config.LoadTools(configMain.General.ToolFile)
	if err != nil {
		return err
	}

	return nil
}

// run method of application
func run() error {
	defer func() {
		mainLog.Debug("Shutting down")
	}()

	// configure application
	err := configure()
	if err != nil {
		return err
	}

	// create MCP server
	mcpServer := server.NewMCPServer(
		appDisplayName,
		appVersion,
		server.WithDescription(appDescription),
		server.WithInstructions(configMain.MCP.Instructions),
		server.WithToolCapabilities(true),
		server.WithInputSchemaValidation(),
	)

	// create tools
	tools, err := createTools()
	if err != nil {
		return err
	}

	// add all tools to the MCP server
	mcpServer.AddTools(tools...)

	// serve MCP requests based on configuration
	switch configMain.MCP.Transport {
	case config.HTTP:
		err := serveHTTP(mcpServer, false)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case config.HTTPS:
		err := serveHTTP(mcpServer, true)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case config.STDIO:
		mainLog.Info("Starting MCP STDIO server")
		// OS signals are handled
		err := server.ServeStdio(mcpServer)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	default:
		return fmt.Errorf("invalid MCP transport: %s", configMain.MCP.Transport)
	}
}

// entry point for application
func main() {
	// run application
	err := run()

	// log fatal error
	if err != nil {
		mainLog.Error("Fatal error", "message", err)
		os.Exit(1)
	}
	os.Exit(0)
}
