"""The two models every phase compares: a logistic baseline and XGBoost."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import numpy as np
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline, make_pipeline
from sklearn.preprocessing import FunctionTransformer, StandardScaler
from xgboost import XGBClassifier

SEED = 42


def build_baseline() -> Pipeline:
    """Logistic regression on log-scaled counts, the classic JIT baseline."""
    return make_pipeline(
        FunctionTransformer(np.log1p),
        StandardScaler(),
        LogisticRegression(max_iter=1000),
    )


def build_xgboost(params: dict[str, Any]) -> XGBClassifier:
    """Gradient-boosted trees; tree models need no scaling."""
    return XGBClassifier(
        objective="binary:logistic",
        eval_metric="aucpr",
        random_state=SEED,
        n_jobs=-1,
        **params,
    )


@dataclass(frozen=True)
class PlattCalibrator:
    """Maps a model's raw log-odds to a calibrated probability.

    Two numbers, applied as sigmoid(slope * logit + intercept), so the Go
    server can reproduce calibration without any ML library.
    """

    slope: float
    intercept: float

    @classmethod
    def fit(cls, logits: np.ndarray, labels: np.ndarray) -> PlattCalibrator:
        regression = LogisticRegression(C=1e6, max_iter=1000)
        regression.fit(logits.reshape(-1, 1), labels)
        return cls(slope=float(regression.coef_[0][0]), intercept=float(regression.intercept_[0]))

    def apply(self, logits: np.ndarray) -> np.ndarray:
        return 1.0 / (1.0 + np.exp(-(self.slope * logits + self.intercept)))


def logits(probabilities: np.ndarray) -> np.ndarray:
    """Inverse sigmoid, clipped so 0 and 1 stay finite."""
    clipped = np.clip(probabilities, 1e-6, 1 - 1e-6)
    return np.log(clipped / (1 - clipped))
