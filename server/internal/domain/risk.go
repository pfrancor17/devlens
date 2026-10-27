package domain

// RiskLevel buckets a risk probability for people and agents to act on.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// Thresholds between levels. Agents rely on these to decide what to do next,
// so they live in the domain rather than in any model.
const (
	mediumRiskThreshold = 0.35
	highRiskThreshold   = 0.65
)

// LevelFor maps a probability in [0, 1] to a risk level.
func LevelFor(probability float64) RiskLevel {
	switch {
	case probability >= highRiskThreshold:
		return RiskHigh
	case probability >= mediumRiskThreshold:
		return RiskMedium
	default:
		return RiskLow
	}
}

// RiskReason is one feature's contribution to a prediction. A positive impact
// raises the risk; a negative one lowers it.
type RiskReason struct {
	Feature string  `json:"feature"`
	Value   float64 `json:"value"`
	Impact  float64 `json:"impact"`
}

// Prediction is what a risk model returns for a set of features.
type Prediction struct {
	Probability float64      `json:"probability"`
	Model       string       `json:"model"`
	Reasons     []RiskReason `json:"reasons"`
}

// RiskAssessment is the full answer about a pull request.
type RiskAssessment struct {
	Repository string       `json:"repository"`
	Number     int          `json:"number"`
	Title      string       `json:"title"`
	Author     string       `json:"author"`
	URL        string       `json:"url"`
	Risk       float64      `json:"risk"`
	Level      RiskLevel    `json:"level"`
	Model      string       `json:"model"`
	Reasons    []RiskReason `json:"reasons"`
	Features   Features     `json:"features"`
}

// Assess combines a pull request, its features and a model prediction.
func Assess(pr PullRequest, features Features, prediction Prediction) RiskAssessment {
	return RiskAssessment{
		Repository: pr.Repository.String(),
		Number:     pr.Number,
		Title:      pr.Title,
		Author:     pr.Author,
		URL:        pr.URL,
		Risk:       prediction.Probability,
		Level:      LevelFor(prediction.Probability),
		Model:      prediction.Model,
		Reasons:    prediction.Reasons,
		Features:   features,
	}
}
