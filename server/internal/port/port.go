// Package port declares the driven ports of DevLens: the interfaces use cases
// need from the outside world (code repositories, model predictors, result
// stores, vector indexes). Adapters implement them; use cases depend on them.
//
// Ports are added together with the first use case that needs them, never
// ahead of time.
package port

import (
	"context"

	"github.com/pfrancor17/devlens/server/internal/domain"
)

// PullRequestSource reads pull requests from a code host.
type PullRequestSource interface {
	PullRequest(ctx context.Context, repo domain.Repository, number int) (domain.PullRequest, error)
}

// RiskModel predicts the probability that a change introduces a defect.
// The heuristic baseline and the trained ONNX model both implement it.
type RiskModel interface {
	Predict(features domain.Features) domain.Prediction
}
