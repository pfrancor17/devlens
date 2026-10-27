// Package heuristic is the baseline risk model: a hand-weighted logistic score
// over diff features. It gives DevLens a working, explainable answer until the
// trained XGBoost model replaces it behind the same port.
//
// The weights encode well-known defect-prediction findings (larger, more
// scattered changes are riskier; changes that touch tests are safer). They
// are not fitted to data and are reported as such through the model name.
package heuristic

import (
	"math"
	"sort"

	"github.com/pfrancor17/devlens/server/internal/domain"
)

// ModelID names this model in every prediction it makes.
const ModelID = "pr-risk-heuristic@0.1.0"

const bias = -3.0

// term is one weighted input of the logistic score.
type term struct {
	feature   string
	weight    float64
	transform func(float64) float64
}

func logScale(v float64) float64 { return math.Log1p(v) }
func identity(v float64) float64 { return v }

// Line counts and file counts grow over orders of magnitude, so they enter on
// a log scale; entropy and the test flag are already small.
var terms = []term{
	{feature: domain.FeatureLinesAdded, weight: 0.30, transform: logScale},
	{feature: domain.FeatureLinesDeleted, weight: 0.10, transform: logScale},
	{feature: domain.FeatureFilesChanged, weight: 0.25, transform: logScale},
	{feature: domain.FeatureDirectoriesChanged, weight: 0.35, transform: logScale},
	{feature: domain.FeatureChangeEntropy, weight: 0.30, transform: identity},
	{feature: domain.FeatureTestsModified, weight: -0.80, transform: identity},
}

// Model implements port.RiskModel.
type Model struct{}

// Predict returns a probability and each feature's additive contribution to
// the log-odds, sorted by absolute impact.
func (Model) Predict(features domain.Features) domain.Prediction {
	logOdds := bias
	reasons := make([]domain.RiskReason, 0, len(terms))

	for _, t := range terms {
		value := features[t.feature]
		impact := t.weight * t.transform(value)
		logOdds += impact
		reasons = append(reasons, domain.RiskReason{Feature: t.feature, Value: value, Impact: round(impact)})
	}

	sort.SliceStable(reasons, func(i, j int) bool {
		return math.Abs(reasons[i].Impact) > math.Abs(reasons[j].Impact)
	})

	return domain.Prediction{
		Probability: round(sigmoid(logOdds)),
		Model:       ModelID,
		ImpactScale: domain.ImpactLogOdds,
		Reasons:     reasons,
	}
}

func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

func round(v float64) float64 { return math.Round(v*1000) / 1000 }
