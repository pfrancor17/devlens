// Package prrisk implements the use case behind the `score_pull_request` tool:
// read a pull request, extract its features and ask the risk model.
package prrisk

import (
	"context"
	"errors"
	"fmt"

	"github.com/pfrancor17/devlens/server/internal/domain"
	"github.com/pfrancor17/devlens/server/internal/port"
)

// ErrInvalidNumber is returned for pull request numbers below 1.
var ErrInvalidNumber = errors.New("pull request number must be positive")

// Service scores pull requests.
type Service struct {
	source port.PullRequestSource
	model  port.RiskModel
}

// NewService wires the use case to a pull request source and a risk model.
func NewService(source port.PullRequestSource, model port.RiskModel) Service {
	return Service{source: source, model: model}
}

// Score assesses the risk of one pull request.
func (s Service) Score(ctx context.Context, repoRef string, number int) (domain.RiskAssessment, error) {
	repo, err := domain.ParseRepository(repoRef)
	if err != nil {
		return domain.RiskAssessment{}, err
	}
	if number < 1 {
		return domain.RiskAssessment{}, fmt.Errorf("%w, got %d", ErrInvalidNumber, number)
	}

	pr, err := s.source.PullRequest(ctx, repo, number)
	if err != nil {
		return domain.RiskAssessment{}, err
	}

	features := domain.ExtractDiffFeatures(pr)
	return domain.Assess(pr, features, s.model.Predict(features)), nil
}
