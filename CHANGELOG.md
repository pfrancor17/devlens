# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses [Semantic Versioning](https://semver.org/).

## [0.2.0] - 2026-10-26

### Added

- `score_pull_request` MCP tool: reads a GitHub pull request, extracts diff features from the feature contract and returns risk, level, model and per-feature contributions.
- Heuristic baseline risk model (`pr-risk-heuristic@0.1.0`) behind the `RiskModel` port, ready to be replaced by the trained model.
- GitHub adapter with file pagination and optional token (`GITHUB_TOKEN`).
- Claude Code plugin: manifest, marketplace, `.mcp.json`, the `risk-assessor` subagent and the `/devlens:assess-pr` skill.
- ADR-003: subagent architecture where measurements decide and agents explain.

### Changed

- The MCP server now receives its use cases from the composition root in `cmd/devlens`.
- Design document (English and Spanish) describes the agent layer.

## [0.1.0] - 2026-10-26

### Added

- Go MCP server with hexagonal layout and a `ping` tool that reports version and module status.
- Architecture test that keeps the domain free of non-standard-library imports.
- Feature contract (`contracts/features.yaml`) for the PR Risk model.
- Design document (English and Spanish) and ADR-001, ADR-002.
- CI for formatting, vet, tests and build.
