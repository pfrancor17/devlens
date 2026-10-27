# ML pipelines

Python side of DevLens: load labeled data, train a baseline and an improved model, calibrate, gate and export to ONNX for the Go server.

## Setup

Requirements: Python 3.11+ and [uv](https://docs.astral.sh/uv/).

```bash
cd ml
uv sync --extra dev
uv run pytest
```

The tests run the whole pipeline on synthetic data, including the ONNX export, so they pass without downloading anything.

## PR Risk model

```bash
uv run devlens-train --config configs/pr_risk.yaml
```

What it does, in order:

1. **Checks the contract.** Every training feature must exist in [`contracts/features.yaml`](../contracts/features.yaml), the file the Go server reads.
2. **Loads the data** and renames dataset columns to contract names, as `configs/pr_risk.yaml` maps them.
3. **Splits by time:** 70% oldest commits to train, 10% to calibrate, 20% newest to test. A random split would let the model learn from the future it is tested on.
4. **Trains two models:** logistic regression on log-scaled counts (baseline) and XGBoost.
5. **Calibrates** XGBoost with Platt scaling on the calibration slice, so a 0.30 means about 30% of such changes were buggy. Platt is two numbers, which the Go server applies without any ML library.
6. **Evaluates on the test slice:** ROC-AUC, PR-AUC, Brier score, calibration error, and precision and recall at the server's high-risk threshold.
7. **Applies the gates** in the config. A model that fails any gate is not exported and the command exits with code 1.
8. **Exports** `models/<name>.onnx` and `models/<name>.json` (feature order, calibration, metrics), after checking that ONNX and XGBoost give the same probabilities.

The JSON report goes to stdout; the decision goes to stderr.

## Before training on ApacheJIT

[ApacheJIT](https://doi.org/10.5281/zenodo.5907001) (Keshavarz and Nagappan, MSR 2022) has about 106,000 commits from 14 Apache projects, roughly 26% of them bug-inducing, from 2003 to 2019.

1. Download the commit-level CSV from the Zenodo record and save it as `ml/data/apachejit_total.csv`. The `data/` folder is git-ignored. Check the record's license before publishing anything derived from it.
2. Open it in a notebook and confirm the column names in `configs/pr_risk.yaml` (`la`, `ld`, `nf`, `nd`, `ent`, `buggy`, `author_date`). If a mapped column is missing, the pipeline stops and lists the columns the file has.
3. **Confirm the units match the Go server.** This is the step that silently breaks models:
   - `la` and `ld` must be absolute line counts. Some JIT datasets normalize them by file size.
   - `ent` must be entropy in bits over modified lines. Kamei et al. normalize it by log2 of the number of files; the server does not. If the dataset is normalized, align one side and note the decision in an ADR.
   - `author_date` must be Unix seconds (`date_unit: s`) or a date string (`date_unit: null`).
4. Run the training and record the metrics in the README of the repository.

## Conventions

- Datasets never enter git; trained models enter `models/` only when they pass their gates.
- Every model is compared against its baseline on the same temporal split.
- Targets for the gates are set after the first baseline run: measure first, promise later.
