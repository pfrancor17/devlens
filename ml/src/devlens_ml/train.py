"""Train, compare, calibrate, gate and export the PR Risk model.

devlens-train --config configs/pr_risk.yaml
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import UTC, datetime
from pathlib import Path

import pandas as pd

from devlens_ml.config import TrainingConfig, load_config
from devlens_ml.contract import check_against_contract
from devlens_ml.data import DATE, LABEL, load_dataset, temporal_split
from devlens_ml.evaluation import check_gates, evaluate
from devlens_ml.export import export_onnx
from devlens_ml.models import PlattCalibrator, build_baseline, build_xgboost, logits

ML_ROOT = Path(__file__).resolve().parents[2]


def run(config: TrainingConfig, frame: pd.DataFrame, out_dir: Path) -> dict:
    """Run the whole pipeline on a prepared frame and return the report.

    The model is exported only when every gate passes; the report always carries
    the metrics, so a rejected model can still be analyzed.
    """
    check_against_contract(config.features)
    train, calibration, test = temporal_split(frame, config.split)
    x = {
        name: part[config.features].to_numpy(float)
        for name, part in (("train", train), ("calibration", calibration), ("test", test))
    }
    y = {
        name: part[LABEL].to_numpy(int)
        for name, part in (("train", train), ("calibration", calibration), ("test", test))
    }

    baseline = build_baseline().fit(x["train"], y["train"])
    xgb = build_xgboost(config.xgboost).fit(x["train"], y["train"])

    raw_calibration = xgb.predict_proba(x["calibration"])[:, 1]
    calibrator = PlattCalibrator.fit(logits(raw_calibration), y["calibration"])
    calibrated_test = calibrator.apply(logits(xgb.predict_proba(x["test"])[:, 1]))

    baseline_metrics = evaluate(y["test"], baseline.predict_proba(x["test"])[:, 1])
    model_metrics = evaluate(y["test"], calibrated_test)
    failures = check_gates(model_metrics, baseline_metrics, config.gates)

    report = {
        "model": f"{config.model_name}@{datetime.now(UTC):%Y.%m.%d}",
        "trained_at": datetime.now(UTC).isoformat(timespec="seconds"),
        "rows": {
            name: len(part)
            for name, part in (("train", train), ("calibration", calibration), ("test", test))
        },
        "test_period": {"from": str(test[DATE].min().date()), "to": str(test[DATE].max().date())},
        "metrics": {"baseline": baseline_metrics.as_dict(), "model": model_metrics.as_dict()},
        "gates_failed": failures,
    }
    if not failures:
        export_onnx(xgb, config.features, calibrator, report, out_dir, sample=x["test"][:256])
    return report


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--config", type=Path, default=ML_ROOT / "configs" / "pr_risk.yaml")
    parser.add_argument("--out", type=Path, default=ML_ROOT.parent / "models")
    args = parser.parse_args(argv)

    config = load_config(args.config)
    frame = load_dataset(config.dataset, config.features)
    report = run(config, frame, args.out)
    json.dump(report, sys.stdout, indent=2)
    print()
    if report["gates_failed"]:
        print(
            "model rejected, nothing exported:",
            *report["gates_failed"],
            sep="\n- ",
            file=sys.stderr,
        )
        return 1
    print(f"exported {report['model']} to {args.out}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
