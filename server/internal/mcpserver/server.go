// Package mcpserver is the driving adapter that exposes DevLens use cases as
// MCP tools. It translates MCP requests into use-case calls and nothing more:
// no business rules live here.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pfrancor17/devlens/server/internal/domain"
	"github.com/pfrancor17/devlens/server/internal/usecase/health"
)

// HealthChecker is the port the ping tool depends on.
type HealthChecker interface {
	Check() domain.Health
}

// New builds the MCP server and registers every DevLens tool.
func New(version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "devlens", Version: version}, nil)
	registerPing(server, health.NewService(version))
	return server
}

type pingInput struct{}

func registerPing(server *mcp.Server, checker HealthChecker) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ping",
		Description: "Reports that DevLens is running, its version and the state of each module.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ pingInput) (*mcp.CallToolResult, domain.Health, error) {
		return nil, checker.Check(), nil
	})
}
