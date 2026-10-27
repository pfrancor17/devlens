from pathlib import Path

import pytest
from conftest import FEATURES, synthetic_commits

from devlens_ml.config import DatasetConfig, SplitConfig, load_config
from devlens_ml.contract import check_against_contract
from devlens_ml.data import DATE, LABEL, load_dataset, prepare, temporal_split

ML_ROOT = Path(__file__).resolve().parents[1]


def test_load_dataset_renames_to_contract_names_and_sorts_by_date(dataset_config):
    frame = load_dataset(dataset_config, FEATURES)

    assert list(frame.columns) == [*FEATURES, LABEL, DATE]
    assert frame[DATE].is_monotonic_increasing
    assert frame[DATE].min().year == 2012
    assert set(frame[LABEL].unique()) == {0, 1}


def test_missing_columns_are_reported_with_what_the_file_has(dataset_config):
    raw = synthetic_commits(100).drop(columns=["ent"])

    with pytest.raises(KeyError, match=r"missing columns \['ent'\].*it has"):
        prepare(raw, dataset_config, FEATURES)


def test_temporal_split_keeps_the_future_out_of_training(dataset_config):
    frame = load_dataset(dataset_config, FEATURES)
    train, calibration, test = temporal_split(frame, SplitConfig(0.7, 0.1, 0.2))

    assert len(train) + len(calibration) + len(test) == len(frame)
    assert train[DATE].max() <= calibration[DATE].min()
    assert calibration[DATE].max() <= test[DATE].min()


def test_split_fractions_must_sum_to_one():
    with pytest.raises(ValueError):
        SplitConfig(train=0.8, calibration=0.1, test=0.2)


def test_training_features_must_exist_in_the_contract():
    check_against_contract(FEATURES)
    with pytest.raises(ValueError, match="not in features.yaml"):
        check_against_contract([*FEATURES, "lines_of_vibes"])


def test_repository_config_is_valid_against_the_contract():
    config = load_config(ML_ROOT / "configs" / "pr_risk.yaml")

    check_against_contract(config.features)
    assert set(config.features) <= set(config.dataset.columns.values())
    assert config.dataset.path.is_relative_to(ML_ROOT)


def test_labels_must_contain_both_classes(dataset_config: DatasetConfig):
    raw = synthetic_commits(100).assign(buggy=0)

    with pytest.raises(ValueError, match="both 0 and 1"):
        prepare(raw, dataset_config, FEATURES)


def test_rows_with_missing_values_are_dropped(dataset_config):
    raw = synthetic_commits(100)
    raw.loc[0, "la"] = None

    frame = prepare(raw, dataset_config, FEATURES)

    assert len(frame) == 99
    assert not frame.isna().any().any()
