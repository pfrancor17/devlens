"""Shared fixtures: a synthetic commit dataset with a known signal."""

from __future__ import annotations

from pathlib import Path

import numpy as np
import pandas as pd
import pytest

from devlens_ml.config import DatasetConfig, Gates, SplitConfig, TrainingConfig

FEATURES = [
    "lines_added",
    "lines_deleted",
    "files_changed",
    "directories_changed",
    "change_entropy",
]
COLUMNS = {
    "la": "lines_added",
    "ld": "lines_deleted",
    "nf": "files_changed",
    "nd": "directories_changed",
    "ent": "change_entropy",
}


def synthetic_commits(rows: int = 4000, seed: int = 7) -> pd.DataFrame:
    """Commits in ApacheJIT's raw shape where size and spread drive defects."""
    rng = np.random.default_rng(seed)
    la = rng.lognormal(3.5, 1.4, rows).round()
    ld = rng.lognormal(2.0, 1.2, rows).round()
    nf = np.maximum(1, rng.poisson(1 + la / 60))
    nd = np.maximum(1, np.minimum(nf, rng.poisson(1 + nf / 4)))
    ent = np.where(nf > 1, rng.uniform(0.2, 1.0, rows) * np.log2(nf), 0.0)
    log_odds = -3.2 + 0.45 * np.log1p(la) + 0.3 * np.log1p(nd) + 0.25 * ent
    buggy = rng.random(rows) < 1 / (1 + np.exp(-log_odds))
    timestamps = pd.date_range("2012-01-01", periods=rows, freq="6h")
    dates = (timestamps - pd.Timestamp("1970-01-01")) // pd.Timedelta(seconds=1)
    return pd.DataFrame(
        {
            "la": la,
            "ld": ld,
            "nf": nf,
            "nd": nd,
            "ent": ent,
            "buggy": buggy.astype(int),
            "author_date": dates,
            "project": "synthetic",
        }
    )


@pytest.fixture
def dataset_config(tmp_path: Path) -> DatasetConfig:
    path = tmp_path / "commits.csv"
    synthetic_commits().to_csv(path, index=False)
    return DatasetConfig(
        path=path, label="buggy", date="author_date", columns=COLUMNS, date_unit="s"
    )


@pytest.fixture
def training_config(dataset_config: DatasetConfig) -> TrainingConfig:
    return TrainingConfig(
        model_name="pr-risk-xgb-test",
        dataset=dataset_config,
        features=FEATURES,
        split=SplitConfig(train=0.7, calibration=0.1, test=0.2),
        gates=Gates(min_roc_auc=0.6, must_beat_baseline=False, max_calibration_error=0.15),
        xgboost={"n_estimators": 80, "max_depth": 3, "learning_rate": 0.1},
    )
