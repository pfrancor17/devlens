"""The feature contract shared with the Go server (contracts/features.yaml)."""

from __future__ import annotations

from pathlib import Path

import yaml

DEFAULT_CONTRACT = Path(__file__).resolve().parents[3] / "contracts" / "features.yaml"


def contract_features(path: Path = DEFAULT_CONTRACT) -> set[str]:
    """Names of every feature the contract defines."""
    raw = yaml.safe_load(path.read_text())
    return {feature["name"] for feature in raw["features"]}


def check_against_contract(features: list[str], path: Path = DEFAULT_CONTRACT) -> None:
    """Fail when a training feature is not defined in the contract, or repeats.

    A feature the server cannot name is a feature the server cannot compute, so
    the model would be unusable in production.
    """
    unknown = sorted(set(features) - contract_features(path))
    if unknown:
        raise ValueError(f"features not in {path.name}: {unknown}")
    if len(set(features)) != len(features):
        raise ValueError(f"features repeat: {features}")
