package domain_test

import (
	"math"
	"testing"

	"github.com/pfrancor17/devlens/server/internal/domain"
)

func TestExtractDiffFeatures(t *testing.T) {
	pr := domain.PullRequest{Files: []domain.ChangedFile{
		{Path: "internal/payments/gateway.go", Additions: 30, Deletions: 10},
		{Path: "internal/payments/gateway_test.go", Additions: 20, Deletions: 0},
		{Path: "cmd/api/main.go", Additions: 0, Deletions: 0},
	}}

	got := domain.ExtractDiffFeatures(pr)

	want := domain.Features{
		domain.FeatureLinesAdded:         50,
		domain.FeatureLinesDeleted:       10,
		domain.FeatureFilesChanged:       3,
		domain.FeatureDirectoriesChanged: 2,
		domain.FeatureTestsModified:      1,
	}
	for name, value := range want {
		if got[name] != value {
			t.Errorf("%s = %v, want %v", name, got[name], value)
		}
	}

	// 40 and 20 modified lines: p = 2/3 and 1/3.
	wantEntropy := -(2.0/3)*math.Log2(2.0/3) - (1.0/3)*math.Log2(1.0/3)
	if math.Abs(got[domain.FeatureChangeEntropy]-wantEntropy) > 1e-9 {
		t.Errorf("change_entropy = %v, want %v", got[domain.FeatureChangeEntropy], wantEntropy)
	}
}

func TestChangeEntropyIsZeroForSingleFileOrEmptyChange(t *testing.T) {
	cases := map[string][]domain.ChangedFile{
		"empty":       nil,
		"single file": {{Path: "a.go", Additions: 100}},
	}
	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			got := domain.ExtractDiffFeatures(domain.PullRequest{Files: files})
			if got[domain.FeatureChangeEntropy] != 0 {
				t.Errorf("change_entropy = %v, want 0", got[domain.FeatureChangeEntropy])
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	cases := map[string]bool{
		"internal/domain/risk_test.go":            true,
		"src/test/java/com/acme/PaymentTest.java": true,
		"tests/test_features.py":                  true,
		"web/src/button.spec.ts":                  true,
		"web/src/__tests__/button.tsx":            true,
		"internal/domain/risk.go":                 false,
		"docs/testing-guide.md":                   false,
		"internal/attestation/signer.go":          false,
		"internal/release/latest.go":              false,
		"app/src/main/kotlin/PaymentTests.kt":     true,
	}
	for filePath, want := range cases {
		if got := domain.IsTestFile(filePath); got != want {
			t.Errorf("IsTestFile(%q) = %v, want %v", filePath, got, want)
		}
	}
}

func TestParseRepository(t *testing.T) {
	repo, err := domain.ParseRepository(" pfrancor17/devlens ")
	if err != nil || repo.String() != "pfrancor17/devlens" {
		t.Fatalf("ParseRepository = %v, %v; want pfrancor17/devlens", repo, err)
	}
	for _, bad := range []string{"", "devlens", "/devlens", "owner/", "a/b/c"} {
		if _, err := domain.ParseRepository(bad); err == nil {
			t.Errorf("ParseRepository(%q) succeeded, want error", bad)
		}
	}
}

func TestLevelFor(t *testing.T) {
	cases := map[float64]domain.RiskLevel{
		0:    domain.RiskLow,
		0.34: domain.RiskLow,
		0.35: domain.RiskMedium,
		0.64: domain.RiskMedium,
		0.65: domain.RiskHigh,
		1:    domain.RiskHigh,
	}
	for probability, want := range cases {
		if got := domain.LevelFor(probability); got != want {
			t.Errorf("LevelFor(%v) = %s, want %s", probability, got, want)
		}
	}
}
