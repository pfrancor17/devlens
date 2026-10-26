// Package domain holds DevLens entities and business rules.
//
// It depends on nothing outside the standard library: no MCP SDK, no GitHub
// client, no ML runtime. Every other layer depends on this one, never the
// other way around.
package domain

// ModuleState describes how far a DevLens module has progressed.
type ModuleState string

const (
	ModulePlanned    ModuleState = "planned"
	ModuleInProgress ModuleState = "in_progress"
	ModuleAvailable  ModuleState = "available"
)

// Module is one stage of the delivery lifecycle that DevLens measures.
type Module struct {
	Name  string      `json:"name"`
	Stage string      `json:"stage"`
	State ModuleState `json:"state"`
}

// Health is the server's self-report: who it is and what it can do today.
type Health struct {
	Service string   `json:"service"`
	Version string   `json:"version"`
	Status  string   `json:"status"`
	Modules []Module `json:"modules"`
}

// Roadmap returns the five DevLens modules in delivery order.
func Roadmap() []Module {
	return []Module{
		{Name: "pr-risk-scorer", Stage: "review", State: ModuleInProgress},
		{Name: "senior-reviewer", Stage: "implement", State: ModulePlanned},
		{Name: "requirement-analyzer", Stage: "analyze", State: ModulePlanned},
		{Name: "implementation-planner", Stage: "plan", State: ModulePlanned},
		{Name: "learning-loop", Stage: "learn", State: ModulePlanned},
	}
}
