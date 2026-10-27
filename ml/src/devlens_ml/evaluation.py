"""Metrics and gates for an imbalanced defect-prediction problem."""

from __future__ import annotations

from dataclasses import asdict, dataclass

import numpy as np
from sklearn.metrics import (
    average_precision_score,
    brier_score_loss,
    precision_score,
    recall_score,
    roc_auc_score,
)

from devlens_ml.config import Gates

# Same cut as domain.LevelFor in the Go server: at or above it, a change is high risk.
HIGH_RISK_THRESHOLD = 0.65


@dataclass(frozen=True)
class Metrics:
    roc_auc: float
    pr_auc: float
    brier: float
    calibration_error: float
    precision_high: float
    recall_high: float
    positive_rate: float

    def as_dict(self) -> dict[str, float]:
        return {name: round(value, 4) for name, value in asdict(self).items()}


def evaluate(labels: np.ndarray, probabilities: np.ndarray) -> Metrics:
    """Score probabilities against labels.

    PR-AUC and recall matter more than accuracy here: with about one buggy
    commit in four, a model that always says "clean" is 75% accurate and useless.
    """
    flagged = probabilities >= HIGH_RISK_THRESHOLD
    return Metrics(
        roc_auc=float(roc_auc_score(labels, probabilities)),
        pr_auc=float(average_precision_score(labels, probabilities)),
        brier=float(brier_score_loss(labels, probabilities)),
        calibration_error=expected_calibration_error(labels, probabilities),
        precision_high=float(precision_score(labels, flagged, zero_division=0)),
        recall_high=float(recall_score(labels, flagged, zero_division=0)),
        positive_rate=float(labels.mean()),
    )


def expected_calibration_error(
    labels: np.ndarray, probabilities: np.ndarray, bins: int = 10
) -> float:
    """Average gap between predicted and observed defect rates, weighted by bin size."""
    edges = np.linspace(0, 1, bins + 1)
    bin_ids = np.clip(np.digitize(probabilities, edges[1:-1]), 0, bins - 1)
    error = 0.0
    for b in range(bins):
        in_bin = bin_ids == b
        if in_bin.any():
            gap = abs(probabilities[in_bin].mean() - labels[in_bin].mean())
            error += gap * in_bin.mean()
    return float(error)


def check_gates(model: Metrics, baseline: Metrics, gates: Gates) -> list[str]:
    """Return the gates the model fails; an empty list means it may ship."""
    failures = []
    if model.roc_auc < gates.min_roc_auc:
        failures.append(f"ROC-AUC {model.roc_auc:.3f} below the floor {gates.min_roc_auc}")
    if gates.must_beat_baseline and model.pr_auc <= baseline.pr_auc:
        failures.append(
            f"PR-AUC {model.pr_auc:.3f} does not beat the baseline's {baseline.pr_auc:.3f}"
        )
    if model.calibration_error > gates.max_calibration_error:
        failures.append(
            f"calibration error {model.calibration_error:.3f} above {gates.max_calibration_error}"
        )
    return failures
