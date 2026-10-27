# ADR-003: DevLens works through specialized subagents, decided by measurements

- **Status:** Accepted
- **Date:** 2026-10-26

## Context

DevLens exposes measurements (risk, ambiguity, duplication, impact) as MCP tools. Tools alone leave every decision to whatever general agent calls them: it must know which tool to use at each stage, how to read the numbers and what to do with them. That knowledge gets repeated in every prompt, drifts between sessions, and turns measurements back into opinions.

The delivery lifecycle DevLens covers has five distinct stages, each with its own inputs, tools and expertise. A single agent holding all of them in context carries a lot of material it does not need for the task at hand.

## Decision

DevLens ships as a **Claude Code plugin** with three layers:

| Layer | Lives in | Responsibility |
| --- | --- | --- |
| Measurement | Go MCP server (`server/`) | Compute numbers with ML models. No judgment |
| Judgment | Subagents (`agents/`) | Read the numbers, apply a written decision policy, turn top reasons into actions |
| Workflow | Skills (`skills/`) | Entry points for people (`/devlens:assess-pr`) that route work to the right subagent |

Rules:

- **One subagent per module**, each restricted through `tools:` to its own module's MCP tools plus read-only code access: `requirement-analyst`, `planner`, `senior-reviewer`, `risk-assessor`, `learning-analyst`.
- **An orchestrator subagent (`devlens-lead`)** arrives once at least two specialists exist. It chains stages and stops the flow when a measurement says so (for example, high ambiguity returns questions instead of a plan).
- **Measurements decide; agents explain.** Every subagent has a decision policy table keyed on the model's level. A subagent never moves an item to a different level than the model gave it.
- **Thresholds live in the domain** (`domain.LevelFor`), not in prompts, so changing them is a code change with tests.
- **Every answer names the model version**, so a heuristic baseline is never mistaken for a trained model.

## Consequences

**Positive**

- Each subagent works with a small, focused context and a narrow tool set.
- Decisions are reproducible and auditable: the same measurement leads to the same verdict.
- Models can be replaced (heuristic → XGBoost → DL) without touching agent prompts.
- The plugin installs in one step from this repository.

**Negative**

- Agent prompts are a second artifact to maintain next to the code. Mitigated by keeping policies short and tied to domain levels.
- Subagents depend on Claude Code's plugin and subagent features. Mitigated by keeping the MCP server usable on its own from any MCP client.
- Running the server with `go run` needs Go on the user's machine. Accepted for now; prebuilt binaries can come with the first release.
