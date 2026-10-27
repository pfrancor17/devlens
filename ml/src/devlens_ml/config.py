"""Typed view of a training configuration file."""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml


@dataclass(frozen=True)
class DatasetConfig:
    path: Path
    label: str
    date: str
    columns: dict[str, str]
    date_unit: str | None = None


@dataclass(frozen=True)
class SplitConfig:
    train: float
    calibration: float
    test: float

    def __post_init__(self) -> None:
        total = self.train + self.calibration + self.test
        if abs(total - 1.0) > 1e-9 or min(self.train, self.calibration, self.test) <= 0:
            raise ValueError(f"split fractions must be positive and sum to 1, got {total}")


@dataclass(frozen=True)
class Gates:
    min_roc_auc: float
    must_beat_baseline: bool
    max_calibration_error: float


@dataclass(frozen=True)
class TrainingConfig:
    model_name: str
    dataset: DatasetConfig
    features: list[str]
    split: SplitConfig
    gates: Gates
    xgboost: dict[str, Any] = field(default_factory=dict)


def load_config(path: Path) -> TrainingConfig:
    """Read a YAML config. Relative dataset paths resolve against the ml/ folder."""
    raw = yaml.safe_load(path.read_text())
    ml_root = path.resolve().parent.parent
    dataset = raw["dataset"]
    return TrainingConfig(
        model_name=raw["model_name"],
        dataset=DatasetConfig(
            path=(ml_root / dataset["path"]).resolve(),
            label=dataset["label"],
            date=dataset["date"],
            columns=dict(dataset["columns"]),
            date_unit=dataset.get("date_unit"),
        ),
        features=list(raw["features"]),
        split=SplitConfig(**raw["split"]),
        gates=Gates(**raw["gates"]),
        xgboost=dict(raw.get("xgboost", {})),
    )
