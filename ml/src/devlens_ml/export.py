"""Export a trained model to ONNX with the metadata the Go server needs."""

from __future__ import annotations

import json
from pathlib import Path

import numpy as np
import onnxruntime as ort
from onnxmltools.convert import convert_xgboost
from onnxmltools.convert.common.data_types import FloatTensorType
from xgboost import XGBClassifier

from devlens_ml.models import PlattCalibrator


def export_onnx(
    model: XGBClassifier,
    features: list[str],
    calibrator: PlattCalibrator,
    metadata: dict,
    out_dir: Path,
    sample: np.ndarray,
) -> Path:
    """Write <model>.onnx and <model>.json, and check ONNX matches XGBoost.

    The JSON carries what ONNX cannot: feature order, calibration and metrics.
    """
    out_dir.mkdir(parents=True, exist_ok=True)
    name = metadata["model"]

    onnx_model = convert_xgboost(
        model,
        initial_types=[("features", FloatTensorType([None, len(features)]))],
        target_opset=15,
    )
    onnx_path = out_dir / f"{name}.onnx"
    onnx_path.write_bytes(onnx_model.SerializeToString())

    _check_parity(model, onnx_path, sample)

    meta = {
        **metadata,
        "features": features,
        "input": "features",
        "calibration": {
            "method": "platt",
            "slope": calibrator.slope,
            "intercept": calibrator.intercept,
        },
    }
    (out_dir / f"{name}.json").write_text(json.dumps(meta, indent=2) + "\n")
    return onnx_path


def _check_parity(model: XGBClassifier, onnx_path: Path, sample: np.ndarray) -> None:
    """Fail the export if ONNX and XGBoost disagree on the same rows."""
    session = ort.InferenceSession(str(onnx_path), providers=["CPUExecutionProvider"])
    onnx_probabilities = session.run(None, {"features": sample.astype(np.float32)})[1]
    onnx_positive = (
        np.array([row[1] for row in onnx_probabilities])
        if isinstance(onnx_probabilities, list)
        else onnx_probabilities[:, 1]
    )
    expected = model.predict_proba(sample)[:, 1]
    if not np.allclose(onnx_positive, expected, atol=1e-4):
        worst = float(np.max(np.abs(onnx_positive - expected)))
        raise RuntimeError(f"ONNX export disagrees with XGBoost by up to {worst:.6f}")
