import pytest
from unittest.mock import MagicMock
import numpy as np
from detector import FraudDetector

@pytest.fixture
def mock_feature_store():
    return MagicMock()

@pytest.fixture
def detector(mock_feature_store):
    return FraudDetector(mock_feature_store)

def test_detect_fraud_normal(detector, mock_feature_store):
    # Mock normal features: low velocity, low ratio, normal amount
    mock_feature_store.get_features_vector.return_value = [1, 1.0, 100]
    
    risk_score, is_fraudulent, reason = detector.detect_fraud("tx1", "user1", 100)
    
    assert risk_score < 0.65
    assert not is_fraudulent
    assert "Normal transaction" in reason

def test_detect_fraud_high_velocity(detector, mock_feature_store):
    # Mock high velocity features
    mock_feature_store.get_features_vector.return_value = [10, 1.0, 100]
    
    risk_score, is_fraudulent, reason = detector.detect_fraud("tx2", "user2", 100)
    
    assert "High transaction velocity" in reason

def test_detect_fraud_high_amount_ratio(detector, mock_feature_store):
    # Mock high amount ratio features
    mock_feature_store.get_features_vector.return_value = [1, 5.0, 500]
    
    risk_score, is_fraudulent, reason = detector.detect_fraud("tx3", "user3", 500)
    
    assert "Suspicious amount variance" in reason

def test_verify_security_allow(detector):
    is_safe, risk_score, reason, action = detector.verify_security(0, "entity1", 100) # 0 is generic/unspecified
    assert is_safe
    assert action == "ALLOW"

def test_verify_security_deny(detector):
    # Very high amount should trigger DENY or CHALLENGE
    # amount > 50000 gives 0.5 risk, + other factors
    # In verify_security, it's hardcoded logic
    # Withdraw > 5000 adds 0.2
    # Amount > 50000 adds 0.5
    # Let's check logic:
    # 62: if amount > 10000: risk_score += 0.3
    # 64: if amount > 50000: risk_score += 0.5, action = "CHALLENGE"
    # 69: if domain == WITHDRAW and amount > 5000: risk_score += 0.2
    # 81: if risk_score > 0.8: action = "DENY"
    
    # Withdrawal of 60000 -> 0.3 + 0.5 + 0.2 = 1.0 risk -> DENY
    from ai_security import ai_security_pb2
    is_safe, risk_score, reason, action = detector.verify_security(ai_security_pb2.WITHDRAW, "user1", 60000)
    
    assert not is_safe
    assert action == "DENY"
    assert risk_score >= 1.0
