package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pfrancor17/devlens/server/internal/domain"
	"github.com/pfrancor17/devlens/server/internal/mcpserver"
	"github.com/pfrancor17/devlens/server/internal/usecase/health"
)

type stubScorer struct {
	assessment domain.RiskAssessment
	err        error
}

func (s stubScorer) Score(context.Context, string, int) (domain.RiskAssessment, error) {
	return s.assessment, s.err
}

// connect starts the server in memory and returns a connected client session.
func connect(t *testing.T, scorer mcpserver.PullRequestScorer) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	server := mcpserver.New("v0.1.0", mcpserver.UseCases{
		Health: health.NewService("v0.1.0"),
		PRRisk: scorer,
	})
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

// callTool calls a tool and decodes its structured result into out.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) *mcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if out != nil && !result.IsError {
		raw, err := json.Marshal(result.StructuredContent)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("decode %s result: %v", name, err)
		}
	}
	return result
}

func TestPingToolReturnsHealth(t *testing.T) {
	session := connect(t, stubScorer{})

	var got domain.Health
	callTool(t, session, "ping", map[string]any{}, &got)

	if got.Service != "devlens" || got.Status != "ok" {
		t.Errorf("health = %+v, want service devlens with status ok", got)
	}
}

func TestScorePullRequestToolReturnsAssessment(t *testing.T) {
	pr := domain.PullRequest{
		Repository: domain.Repository{Owner: "acme", Name: "shop"},
		Number:     7,
		Files:      []domain.ChangedFile{{Path: "refund.go", Additions: 40}},
	}
	prediction := domain.Prediction{
		Probability: 0.72,
		Model:       "fixed@test",
		Reasons:     []domain.RiskReason{{Feature: domain.FeatureLinesAdded, Value: 40, Impact: 1.1}},
	}
	session := connect(t, stubScorer{assessment: domain.Assess(pr, domain.ExtractDiffFeatures(pr), prediction)})

	var got domain.RiskAssessment
	callTool(t, session, "score_pull_request", map[string]any{"repo": "acme/shop", "number": 7}, &got)

	if got.Risk != 0.72 || got.Level != domain.RiskHigh {
		t.Errorf("assessment = %+v, want risk 0.72 high", got)
	}
}

func TestScorePullRequestToolReportsErrorsToTheAgent(t *testing.T) {
	session := connect(t, stubScorer{err: errors.New("GitHub answered 404 Not Found")})

	result := callTool(t, session, "score_pull_request", map[string]any{"repo": "acme/shop", "number": 9}, nil)

	if !result.IsError {
		t.Fatalf("IsError = false, want a tool error the agent can read")
	}
}
