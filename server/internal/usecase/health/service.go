// Package health implements the use case behind the `ping` tool.
package health

import "github.com/pfrancor17/devlens/server/internal/domain"

const serviceName = "devlens"

// Service reports the server's health and the state of each module.
type Service struct {
	version string
}

// NewService builds a health service for the given release version.
func NewService(version string) Service {
	return Service{version: version}
}

// Check returns the current health report.
func (s Service) Check() domain.Health {
	return domain.Health{
		Service: serviceName,
		Version: s.version,
		Status:  "ok",
		Modules: domain.Roadmap(),
	}
}
