// Command devlens starts the DevLens MCP server over stdio.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pfrancor17/devlens/server/internal/mcpserver"
)

// version is overridden at build time: go build -ldflags "-X main.version=v0.2.0"
var version = "v0.1.0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// stdout carries the MCP protocol, so logs must go to stderr.
	logger := log.New(os.Stderr, "devlens: ", log.LstdFlags)

	if err := mcpserver.New(version).Run(ctx, &mcp.StdioTransport{}); err != nil {
		logger.Fatalf("server stopped: %v", err)
	}
}
