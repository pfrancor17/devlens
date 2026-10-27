// Command devlens starts the DevLens MCP server over stdio. It is the
// composition root: the only place that knows every concrete adapter.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pfrancor17/devlens/server/internal/adapter/github"
	"github.com/pfrancor17/devlens/server/internal/adapter/heuristic"
	"github.com/pfrancor17/devlens/server/internal/mcpserver"
	"github.com/pfrancor17/devlens/server/internal/usecase/health"
	"github.com/pfrancor17/devlens/server/internal/usecase/prrisk"
)

// version is overridden at build time: go build -ldflags "-X main.version=v0.2.0"
var version = "v0.2.0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// stdout carries the MCP protocol, so logs must go to stderr.
	logger := log.New(os.Stderr, "devlens: ", log.LstdFlags)

	server := mcpserver.New(version, mcpserver.UseCases{
		Health: health.NewService(version),
		PRRisk: prrisk.NewService(github.NewClient(os.Getenv("GITHUB_TOKEN")), heuristic.Model{}),
	})

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		logger.Fatalf("server stopped: %v", err)
	}
}
