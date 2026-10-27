package domain

import (
	"math"
	"path"
	"strings"
)

// Feature names. They must match contracts/features.yaml exactly, because the
// trained model reads features by name.
const (
	FeatureLinesAdded         = "lines_added"
	FeatureLinesDeleted       = "lines_deleted"
	FeatureFilesChanged       = "files_changed"
	FeatureDirectoriesChanged = "directories_changed"
	FeatureChangeEntropy      = "change_entropy"
	FeatureTestsModified      = "tests_modified"
)

// Features holds the model inputs computed for a change, keyed by feature name.
// Booleans are encoded as 0 or 1 so every model sees plain numbers.
type Features map[string]float64

// ExtractDiffFeatures computes the features that depend only on the diff.
// History features (bug history, file age, author experience) arrive with the
// repository-history adapter.
func ExtractDiffFeatures(pr PullRequest) Features {
	var added, deleted int
	directories := make(map[string]struct{}, len(pr.Files))
	testsModified := false

	for _, file := range pr.Files {
		added += file.Additions
		deleted += file.Deletions
		directories[path.Dir(file.Path)] = struct{}{}
		testsModified = testsModified || IsTestFile(file.Path)
	}

	return Features{
		FeatureLinesAdded:         float64(added),
		FeatureLinesDeleted:       float64(deleted),
		FeatureFilesChanged:       float64(len(pr.Files)),
		FeatureDirectoriesChanged: float64(len(directories)),
		FeatureChangeEntropy:      changeEntropy(pr.Files),
		FeatureTestsModified:      boolToFloat(testsModified),
	}
}

// changeEntropy is the Shannon entropy, in bits, of modified lines across
// files. A change concentrated in one file scores 0; a change spread evenly
// over n files scores log2(n).
func changeEntropy(files []ChangedFile) float64 {
	total := 0
	for _, file := range files {
		total += file.Lines()
	}
	if total == 0 {
		return 0
	}

	entropy := 0.0
	for _, file := range files {
		if file.Lines() == 0 {
			continue
		}
		p := float64(file.Lines()) / float64(total)
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// IsTestFile reports whether a path looks like a test file in common
// ecosystems (Go, Java, Python, JavaScript/TypeScript).
func IsTestFile(filePath string) bool {
	base := path.Base(filePath)
	stem := strings.TrimSuffix(base, path.Ext(base))
	lowerBase := strings.ToLower(base)

	switch {
	case strings.HasSuffix(lowerBase, "_test.go"),
		strings.HasSuffix(lowerBase, ".py") && (strings.HasPrefix(lowerBase, "test_") || strings.HasSuffix(stem, "_test")),
		strings.Contains(lowerBase, ".test."),
		strings.Contains(lowerBase, ".spec."),
		strings.HasSuffix(stem, "Test"), strings.HasSuffix(stem, "Tests"): // Java and Kotlin convention
		return true
	}
	for _, dir := range strings.Split(strings.ToLower(path.Dir(filePath)), "/") {
		if dir == "test" || dir == "tests" || dir == "__tests__" {
			return true
		}
	}
	return false
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
