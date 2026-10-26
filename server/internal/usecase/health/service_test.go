package health_test

import (
	"testing"

	"github.com/pfrancor17/devlens/server/internal/domain"
	"github.com/pfrancor17/devlens/server/internal/usecase/health"
)

func TestCheckReportsVersionAndAllModules(t *testing.T) {
	got := health.NewService("v0.1.0").Check()

	if got.Status != "ok" {
		t.Errorf("Status = %q, want %q", got.Status, "ok")
	}
	if got.Version != "v0.1.0" {
		t.Errorf("Version = %q, want %q", got.Version, "v0.1.0")
	}
	if len(got.Modules) != len(domain.Roadmap()) {
		t.Errorf("Modules = %d, want %d", len(got.Modules), len(domain.Roadmap()))
	}
}
