package prrisk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pfrancor17/devlens/server/internal/domain"
	"github.com/pfrancor17/devlens/server/internal/usecase/prrisk"
)

type fakeSource struct {
	pr  domain.PullRequest
	err error
}

func (f fakeSource) PullRequest(context.Context, domain.Repository, int) (domain.PullRequest, error) {
	return f.pr, f.err
}

type fixedModel struct{ probability float64 }

func (m fixedModel) Predict(domain.Features) domain.Prediction {
	return domain.Prediction{Probability: m.probability, Model: "fixed@test"}
}

func TestScoreCombinesPullRequestFeaturesAndPrediction(t *testing.T) {
	pr := domain.PullRequest{
		Repository: domain.Repository{Owner: "acme", Name: "shop"},
		Number:     7,
		Title:      "Add refunds",
		Files:      []domain.ChangedFile{{Path: "refund.go", Additions: 40}},
	}
	service := prrisk.NewService(fakeSource{pr: pr}, fixedModel{probability: 0.8})

	got, err := service.Score(context.Background(), "acme/shop", 7)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	if got.Repository != "acme/shop" || got.Number != 7 || got.Title != "Add refunds" {
		t.Errorf("identity = %+v", got)
	}
	if got.Risk != 0.8 || got.Level != domain.RiskHigh || got.Model != "fixed@test" {
		t.Errorf("risk = %v %s %s, want 0.8 high fixed@test", got.Risk, got.Level, got.Model)
	}
	if got.Features[domain.FeatureLinesAdded] != 40 {
		t.Errorf("lines_added = %v, want 40", got.Features[domain.FeatureLinesAdded])
	}
}

func TestScoreRejectsInvalidInputBeforeCallingGitHub(t *testing.T) {
	service := prrisk.NewService(fakeSource{err: errors.New("must not be called")}, fixedModel{})

	if _, err := service.Score(context.Background(), "not-a-repo", 1); !errors.Is(err, domain.ErrInvalidRepository) {
		t.Errorf("bad repo: err = %v, want ErrInvalidRepository", err)
	}
	if _, err := service.Score(context.Background(), "acme/shop", 0); !errors.Is(err, prrisk.ErrInvalidNumber) {
		t.Errorf("bad number: err = %v, want ErrInvalidNumber", err)
	}
}

func TestScorePropagatesSourceErrors(t *testing.T) {
	sourceErr := errors.New("rate limited")
	service := prrisk.NewService(fakeSource{err: sourceErr}, fixedModel{})

	if _, err := service.Score(context.Background(), "acme/shop", 1); !errors.Is(err, sourceErr) {
		t.Errorf("err = %v, want %v", err, sourceErr)
	}
}
