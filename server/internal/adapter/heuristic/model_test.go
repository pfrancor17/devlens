package heuristic_test

import (
	"testing"

	"github.com/pfrancor17/devlens/server/internal/adapter/heuristic"
	"github.com/pfrancor17/devlens/server/internal/domain"
)

func features(added, files, dirs float64, entropy float64, tests bool) domain.Features {
	testsFlag := 0.0
	if tests {
		testsFlag = 1
	}
	return domain.Features{
		domain.FeatureLinesAdded:         added,
		domain.FeatureFilesChanged:       files,
		domain.FeatureDirectoriesChanged: dirs,
		domain.FeatureChangeEntropy:      entropy,
		domain.FeatureTestsModified:      testsFlag,
	}
}

func TestSmallFocusedChangeWithTestsIsLowRisk(t *testing.T) {
	got := heuristic.Model{}.Predict(features(15, 2, 1, 0.5, true))

	if level := domain.LevelFor(got.Probability); level != domain.RiskLow {
		t.Errorf("level = %s (p=%v), want low", level, got.Probability)
	}
}

func TestLargeScatteredChangeWithoutTestsIsHighRisk(t *testing.T) {
	got := heuristic.Model{}.Predict(features(1200, 40, 15, 4.5, false))

	if level := domain.LevelFor(got.Probability); level != domain.RiskHigh {
		t.Errorf("level = %s (p=%v), want high", level, got.Probability)
	}
}

func TestAddingTestsLowersRisk(t *testing.T) {
	without := heuristic.Model{}.Predict(features(300, 8, 4, 2.5, false))
	with := heuristic.Model{}.Predict(features(300, 8, 4, 2.5, true))

	if with.Probability >= without.Probability {
		t.Errorf("risk with tests %v >= without tests %v", with.Probability, without.Probability)
	}
}

func TestReasonsAreSortedByAbsoluteImpactAndNameTheModel(t *testing.T) {
	got := heuristic.Model{}.Predict(features(300, 8, 4, 2.5, true))

	if got.Model != heuristic.ModelID {
		t.Errorf("model = %q, want %q", got.Model, heuristic.ModelID)
	}
	for i := 1; i < len(got.Reasons); i++ {
		prev, cur := got.Reasons[i-1].Impact, got.Reasons[i].Impact
		if abs(prev) < abs(cur) {
			t.Fatalf("reasons not sorted by |impact|: %+v", got.Reasons)
		}
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
