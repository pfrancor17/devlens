import numpy as np

from devlens_ml.config import Gates
from devlens_ml.evaluation import Metrics, check_gates, evaluate, expected_calibration_error
from devlens_ml.models import PlattCalibrator, logits


def metrics(roc_auc=0.75, pr_auc=0.5, calibration_error=0.03) -> Metrics:
    return Metrics(
        roc_auc=roc_auc,
        pr_auc=pr_auc,
        brier=0.18,
        calibration_error=calibration_error,
        precision_high=0.5,
        recall_high=0.3,
        positive_rate=0.25,
    )


GATES = Gates(min_roc_auc=0.65, must_beat_baseline=True, max_calibration_error=0.1)


def test_a_perfectly_calibrated_model_has_no_calibration_error():
    probabilities = np.repeat([0.1, 0.5, 0.9], 1000)
    labels = np.concatenate([np.r_[np.ones(n), np.zeros(1000 - n)] for n in (100, 500, 900)])

    assert expected_calibration_error(labels, probabilities) < 1e-9


def test_an_overconfident_model_has_calibration_error():
    labels = np.r_[np.ones(500), np.zeros(500)]

    assert abs(expected_calibration_error(labels, np.full(1000, 0.9)) - 0.4) < 1e-9


def test_evaluate_rewards_ranking_buggy_commits_first():
    labels = np.r_[np.ones(100), np.zeros(300)]
    perfect = evaluate(labels, np.r_[np.full(100, 0.9), np.full(300, 0.1)])

    assert perfect.roc_auc == 1.0 and perfect.recall_high == 1.0 and perfect.positive_rate == 0.25


def test_gates_pass_a_model_that_beats_the_baseline():
    assert check_gates(metrics(pr_auc=0.55), metrics(pr_auc=0.45), GATES) == []


def test_gates_name_every_failure():
    failures = check_gates(
        metrics(roc_auc=0.6, pr_auc=0.4, calibration_error=0.2), metrics(pr_auc=0.45), GATES
    )

    assert len(failures) == 3
    assert any("ROC-AUC" in f for f in failures)
    assert any("baseline" in f for f in failures)
    assert any("calibration" in f for f in failures)


def test_platt_calibration_fixes_an_overconfident_score():
    rng = np.random.default_rng(1)
    true_probability = rng.uniform(0.05, 0.6, 20000)
    labels = (rng.random(20000) < true_probability).astype(int)
    overconfident = 1 / (1 + np.exp(-3 * logits(true_probability)))

    calibrator = PlattCalibrator.fit(logits(overconfident), labels)
    calibrated = calibrator.apply(logits(overconfident))

    assert expected_calibration_error(labels, calibrated) < 0.02
    assert expected_calibration_error(labels, overconfident) > 0.05
