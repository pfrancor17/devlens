import json
from dataclasses import replace

import numpy as np
import onnxruntime as ort
from conftest import FEATURES

from devlens_ml.data import load_dataset
from devlens_ml.models import PlattCalibrator
from devlens_ml.train import run


def test_pipeline_trains_gates_and_exports_a_matching_onnx_model(training_config, tmp_path):
    frame = load_dataset(training_config.dataset, FEATURES)

    report = run(training_config, frame, tmp_path)

    model = report["metrics"]["model"]
    baseline = report["metrics"]["baseline"]
    assert model["roc_auc"] > 0.6 and baseline["roc_auc"] > 0.6
    assert report["gates_failed"] == []

    name = report["model"]
    meta = json.loads((tmp_path / f"{name}.json").read_text())
    assert meta["features"] == FEATURES
    assert meta["calibration"]["method"] == "platt"

    # The Go server will feed features in this order and apply Platt itself.
    session = ort.InferenceSession(
        str(tmp_path / f"{name}.onnx"), providers=["CPUExecutionProvider"]
    )
    row = frame[FEATURES].to_numpy(np.float32)[:5]
    outputs = session.run(None, {meta["input"]: row})
    assert len(outputs) == 2
    calibrator = PlattCalibrator(**{k: meta["calibration"][k] for k in ("slope", "intercept")})
    assert np.all((calibrator.apply(np.zeros(1)) > 0) & (calibrator.apply(np.zeros(1)) < 1))


def test_pipeline_refuses_to_export_a_model_that_fails_its_gates(training_config, tmp_path):
    strict = replace(training_config, gates=replace(training_config.gates, min_roc_auc=0.99))
    frame = load_dataset(strict.dataset, FEATURES)

    report = run(strict, frame, tmp_path)

    assert any("ROC-AUC" in failure for failure in report["gates_failed"])
    assert report["metrics"]["model"]["roc_auc"] > 0
    assert not list(tmp_path.glob("*.onnx"))
