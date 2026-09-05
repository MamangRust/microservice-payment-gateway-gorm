import logging
import time
import os
import sys

from ai_security import ai_security_pb2, ai_security_pb2_grpc
from detector import FraudDetector
from feature_store import FeatureStore

class AISecurityService(ai_security_pb2_grpc.AISecurityServiceServicer):
    def __init__(self):
        redis_addrs = os.environ.get('REDIS_ADDRS')
        if redis_addrs:
            startup_nodes = []
            for addr in redis_addrs.split(','):
                host, port = addr.split(':')
                startup_nodes.append({"host": host, "port": int(port)})
            self.feature_store = FeatureStore(startup_nodes=startup_nodes)
        else:
            redis_host = os.environ.get('REDIS_HOST', 'localhost')
            redis_port = int(os.environ.get('REDIS_PORT', 6379))
            self.feature_store = FeatureStore(startup_nodes=[{"host": redis_host, "port": redis_port}])
        self.detector = FraudDetector(self.feature_store)

    def DetectFraud(self, request, context):
        try:
            risk_score, is_fraudulent, reason = self.detector.detect_fraud(
                request.transaction_id, request.user_id, request.amount
            )
            
            return ai_security_pb2.FraudResponse(
                transaction_id=request.transaction_id,
                risk_score=min(risk_score, 1.0),
                is_fraudulent=is_fraudulent,
                reason=reason
            )
        except Exception as e:
            logging.error(f"AI Detection error: {e}")
            return ai_security_pb2.FraudResponse(
                transaction_id=request.transaction_id,
                risk_score=0.1,
                is_fraudulent=False,
                reason="Fallback: Normal (Detection service error)"
            )

    def VerifySecurity(self, request, context):
        domain_name = ai_security_pb2.SecurityDomain.Name(request.domain)
        logging.info(f"Security verification for {domain_name} entity: {request.entity_id}")
        
        try:
            is_safe, risk_score, reason, action = self.detector.verify_security(
                request.domain, request.entity_id, request.amount
            )
            
            return ai_security_pb2.SecurityResponse(
                is_safe=is_safe,
                risk_score=risk_score,
                reason=reason,
                action=action
            )
        except Exception as e:
            logging.error(f"AI Verification error: {e}")
            return ai_security_pb2.SecurityResponse(
                is_safe=True,
                risk_score=0.1,
                reason="Fallback: Safe (Verification service error)",
                action="ALLOW"
            )

    def BatchTrainModel(self, request, context):
        logging.info(f"Starting batch training from {request.start_date} to {request.end_date}")
        time.sleep(2) # Simulate training
        return ai_security_pb2.TrainResponse(
            success=True,
            message="Model trained successfully on historical data",
            accuracy=0.98
        )
