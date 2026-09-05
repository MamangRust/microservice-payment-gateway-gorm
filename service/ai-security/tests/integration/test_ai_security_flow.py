import pytest
import os
import time
import docker
from feature_store import FeatureStore
from detector import FraudDetector
from service import AISecurityService
from ai_security import ai_security_pb2
from redis import Redis

@pytest.fixture(scope="module")
def redis_container():
    """Manual Docker container management to bypass testcontainers bugs."""
    client = docker.from_env()
    image = "redis:7.2"
    try:
        client.images.get(image)
    except docker.errors.ImageNotFound:
        client.images.pull(image)
    
    container = client.containers.run(
        image,
        detach=True,
        ports={'6379/tcp': None}
    )
    
    container.reload()
    host_port = container.ports['6379/tcp'][0]['HostPort']
    
    # Wait for redis
    retries = 10
    while retries > 0:
        try:
            r = Redis(host='localhost', port=int(host_port))
            if r.ping():
                break
        except:
            time.sleep(1)
            retries -= 1
            
    yield {'host': 'localhost', 'port': int(host_port)}
    
    container.stop()
    container.remove()

def test_ai_security_integration_all(redis_container):
    host = redis_container['host']
    port = redis_container['port']
    
    from unittest.mock import patch
    from redis import Redis
    
    # 1. Test FeatureStore directly with patching
    mock_redis = Redis(host=host, port=port, decode_responses=True)
    
    with patch("feature_store.RedisCluster", return_value=mock_redis):
        fs = FeatureStore(startup_nodes=[{"host": host, "port": port}])
        fs.redis = mock_redis
        
        user_id = 7777
        assert fs.get_transaction_velocity(user_id) == 1
        assert fs.get_transaction_velocity(user_id) == 2
        assert fs.get_amount_stats(user_id, 100) == 1.0
        assert fs.get_amount_stats(user_id, 200) == 2.0
        
        # 2. Test Full AISecurityService Flow
        service = AISecurityService()
        service.feature_store = fs
        service.detector = FraudDetector(fs)
        
        request = ai_security_pb2.FraudRequest(
            transaction_id="itx_final",
            user_id=user_id,
            amount=500.0
        )
        
        response = service.DetectFraud(request, None)
        
        assert response.transaction_id == "itx_final"
        # 500 / 150 = 3.33 -> Suspicious variance
        assert "Suspicious amount variance" in response.reason
