# ML pipelines

Python side of DevLens: data, features, training, evaluation and export to ONNX.

| Folder | Purpose |
| --- | --- |
| `pipelines/` | One pipeline per model (`pr_risk`, `clones`, `smells`, `requirement_complexity`, `requirement_ambiguity`) |
| `features/` | Implementation of [`contracts/features.yaml`](../contracts/features.yaml) |
| `evaluation/` | Metrics, temporal splits and minimum thresholds checked in CI |

Conventions:

- Datasets are downloaded by scripts into `ml/data/` and never committed.
- Every run is tracked in MLflow; only released models are copied to `models/`.
- Validation uses temporal splits for any dated data (commits, issues).

The first pipeline (PR Risk) arrives during phase 1.
