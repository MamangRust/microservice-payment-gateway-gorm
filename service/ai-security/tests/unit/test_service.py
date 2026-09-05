import pytest
from unittest.mock import MagicMock, patch
from service import AISecurityService
from ai_security import ai_security_pb2

@pytest.fixture
def mock_detector():
    with patch('service.FraudDetector') as mock:
        yield mock.return_value

@pytest.fixture
def mock_feature_store():
    with patch('service.FeatureStore') as mock:
        yield mock.return_value

@pytest.fixture
def service(mock_detector, mock_feature_store):
    # Mock environment variable for redis host
    with patch.dict('os.environ', {'REDIS_HOST': 'localhost'}):
        return AISecurityService()

def test_detect_fraud_success(service, mock_detector):
    # Mock detector response
    mock_detector.detect_fraud.return_value = (0.5, False, "Normal")
    
    request = ai_security_pb2.FraudRequest(
        transaction_id="tx123",
        user_id=456,
        amount=100.0
    )
    
    response = service.DetectFraud(request, None)
    
    assert response.transaction_id == "tx123"
    assert response.risk_score == 0.5
    assert not response.is_fraudulent
    assert response.reason == "Normal"

def test_detect_fraud_error_fallback(service, mock_detector):
    # Mock detector to raise exception
    mock_detector.detect_fraud.side_effect = Exception("ML Engine failure")
    
    request = ai_security_pb2.FraudRequest(transaction_id="tx_err", user_id=0)
    
    response = service.DetectFraud(request, None)
    
    assert response.transaction_id == "tx_err"
    assert response.risk_score == 0.1
    assert "Fallback" in response.reason

def test_verify_security_success(service, mock_detector):
    mock_detector.verify_security.return_value = (True, 0.2, "Safe", "ALLOW")
    
    request = ai_security_pb2.SecurityRequest(
        domain=ai_security_pb2.CARD,
        entity_id="card123",
        amount=50.0
    )
    
    response = service.VerifySecurity(request, None)
    
    assert response.is_safe
    assert response.action == "ALLOW"
    assert response.reason == "Safe"

def test_batch_train_model(service):
    request = ai_security_pb2.TrainRequest(
        start_date="2024-01-01",
        end_date="2024-01-31"
    )
    
    response = service.BatchTrainModel(request, None)
    
    assert response.success
    assert response.accuracy > 0.9
