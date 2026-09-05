import pytest
from unittest.mock import MagicMock, patch
import json
from feature_store import FeatureStore

@pytest.fixture
def mock_redis():
    with patch('feature_store.RedisCluster') as mock:
        yield mock.return_value

@pytest.fixture
def feature_store(mock_redis):
    # Pass some dummy nodes so it doesn't try to use env vars in a way that might fail
    return FeatureStore(startup_nodes=[{"host": "localhost", "port": 6379}])

def test_get_transaction_velocity(feature_store, mock_redis):
    # Mock return value for zcard
    mock_redis.zcard.return_value = 5
    
    velocity = feature_store.get_transaction_velocity("user123")
    
    assert velocity == 5
    assert mock_redis.zadd.called
    assert mock_redis.zremrangebyscore.called
    assert mock_redis.zcard.called

def test_get_amount_stats_first_time(feature_store, mock_redis):
    # Mock redis.get to return None (first time)
    mock_redis.get.return_value = None
    
    ratio = feature_store.get_amount_stats("user123", 100)
    
    assert ratio == 1.0
    # Should set initial stats
    mock_redis.set.assert_called()
    call_args = mock_redis.set.call_args[0]
    assert "user123" in call_args[0]
    stats = json.loads(call_args[1])
    assert stats["avg"] == 100
    assert stats["count"] == 1

def test_get_amount_stats_second_time(feature_store, mock_redis):
    # Mock redis.get to return existing stats
    mock_redis.get.return_value = json.dumps({"avg": 100, "count": 1})
    
    # Current amount is 200, avg is 100 -> ratio should be 2.0
    ratio = feature_store.get_amount_stats("user123", 200)
    
    assert ratio == 2.0
    # New avg should be (100*1 + 200) / 2 = 150
    mock_redis.set.assert_called()
    stats = json.loads(mock_redis.set.call_args[0][1])
    assert stats["avg"] == 150
    assert stats["count"] == 2

def test_get_features_vector(feature_store, mock_redis):
    # Mock both methods
    with patch.object(FeatureStore, 'get_transaction_velocity', return_value=3), \
         patch.object(FeatureStore, 'get_amount_stats', return_value=1.5):
        
        vector = feature_store.get_features_vector("user123", 150)
        
        assert vector == [3, 1.5, 150]
