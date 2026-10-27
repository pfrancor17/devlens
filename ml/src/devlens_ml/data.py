"""Load a labeled commit dataset into contract feature names, ordered by time."""

from __future__ import annotations

import pandas as pd

from devlens_ml.config import DatasetConfig, SplitConfig

LABEL = "label"
DATE = "date"


def load_dataset(config: DatasetConfig, features: list[str]) -> pd.DataFrame:
    """Read the CSV, rename columns to contract names and sort by date.

    Returns one row per commit with the requested features, `label` (0 or 1)
    and `date`.
    """
    if not config.path.exists():
        raise FileNotFoundError(
            f"dataset not found at {config.path}; download it as ml/README.md explains"
        )
    raw = pd.read_csv(config.path)
    return prepare(raw, config, features)


def prepare(raw: pd.DataFrame, config: DatasetConfig, features: list[str]) -> pd.DataFrame:
    """Rename, validate and order a raw dataset. Separate from I/O for testing."""
    required = [config.label, config.date, *config.columns]
    missing = [column for column in required if column not in raw.columns]
    if missing:
        raise KeyError(f"dataset is missing columns {missing}; it has {sorted(raw.columns)}")

    renamed = raw.rename(columns={**config.columns, config.label: LABEL, config.date: DATE})
    unmapped = [feature for feature in features if feature not in renamed.columns]
    if unmapped:
        raise KeyError(f"features {unmapped} have no column mapped to them in the config")

    frame = renamed[[*features, LABEL, DATE]].copy()
    frame[DATE] = pd.to_datetime(frame[DATE], unit=config.date_unit)
    frame[LABEL] = frame[LABEL].astype(int)
    frame = frame.dropna()

    labels = set(frame[LABEL].unique())
    if not labels <= {0, 1} or len(labels) < 2:
        raise ValueError(f"label must contain both 0 and 1, found {sorted(labels)}")

    return frame.sort_values(DATE, kind="stable").reset_index(drop=True)


def temporal_split(
    frame: pd.DataFrame, split: SplitConfig
) -> tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    """Split a date-sorted frame into train, calibration and test, oldest first.

    A random split would let the model learn from the future it is tested on;
    a temporal one measures what happens when it meets new commits.
    """
    train_end = int(len(frame) * split.train)
    calibration_end = train_end + int(len(frame) * split.calibration)
    return (
        frame.iloc[:train_end],
        frame.iloc[train_end:calibration_end],
        frame.iloc[calibration_end:],
    )
