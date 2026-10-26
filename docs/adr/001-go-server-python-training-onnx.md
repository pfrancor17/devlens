# ADR-001: Go for the MCP server, Python for training, ONNX as the bridge

- **Status:** Accepted
- **Date:** 2026-10-26

## Context

DevLens has two very different workloads:

1. **Serving:** an MCP server that answers agent requests with low latency, calls the GitHub API and runs model inference. It must be easy to install and run as a single binary next to the agent.
2. **Training:** data exploration, feature engineering, model training, evaluation and experiment tracking. This is where the ML ecosystem lives (pandas, scikit-learn, XGBoost, PyTorch, Hugging Face, SHAP, MLflow).

Doing both in one language forces a bad trade-off: Python for serving means shipping an interpreter and heavy dependencies to every user; Go for training means giving up the ML ecosystem.

## Decision

- The **MCP server is written in Go**, using the official MCP Go SDK.
- **Training pipelines are written in Python.**
- Trained models are **exported to ONNX** and loaded by the Go server through an ONNX runtime adapter.
- Features are defined once in `contracts/features.yaml` and implemented in both languages, with contract tests proving they produce the same values.

## Consequences

**Positive**

- The server ships as one static binary with no Python dependency.
- Training uses the best tools available without compromise.
- Models are versioned artifacts with clear boundaries; the server does not care how a model was trained.

**Negative**

- Features are implemented twice. Mitigated by the shared contract and contract tests in CI.
- Not every model or pre-processing step exports cleanly to ONNX. Mitigated by keeping pre-processing in the feature layer and choosing exportable architectures.
- Two toolchains in one repository. Mitigated by separate folders (`server/`, `ml/`) and CI jobs per language.
