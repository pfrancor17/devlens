// Package mcpserver is the driving adapter that exposes DevLens use cases as
// MCP tools. It translates MCP requests into use-case calls and nothing more:
// no business rules live here.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pfrancor17/devlens/server/internal/domain"
)

// HealthChecker is the use case behind the ping tool.
type HealthChecker interface {
	Check() domain.Health
}

// PullRequestScorer is the use case behind the score_pull_request tool.
type PullRequestScorer interface {
	Score(ctx context.Context, repo string, number int) (domain.RiskAssessment, error)
}

// UseCases are the capabilities the server exposes as tools.
type UseCases struct {
	Health HealthChecker
	PRRisk PullRequestScorer
}

// New builds the MCP server and registers every DevLens tool.
func New(version string, useCases UseCases) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "devlens", Version: version}, nil)
	registerPing(server, useCases.Health)
	registerScorePullRequest(server, useCases.PRRisk)
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

type scorePullRequestInput struct {
	Repo   string `json:"repo" jsonschema:"repository as owner/name, for example modelcontextprotocol/go-sdk"`
	Number int    `json:"number" jsonschema:"pull request number"`
}

func registerScorePullRequest(server *mcp.Server, scorer PullRequestScorer) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "score_pull_request",
		Description: "Estimates the probability (0-1) that a GitHub pull request introduces a defect, " +
			"with its level (low, medium, high), the model used and each feature's contribution to the score.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in scorePullRequestInput) (*mcp.CallToolResult, domain.RiskAssessment, error) {
		assessment, err := scorer.Score(ctx, in.Repo, in.Number)
		return nil, assessment, err
	})
}
