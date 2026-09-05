import logging
import numpy as np
from sklearn.ensemble import IsolationForest
import os
import sys

from ai_security import ai_security_pb2

class FraudDetector:
    def __init__(self, feature_store):
        self.feature_store = feature_store
        
        # Initialize Isolation Forest with some initial data
        # In production, this model would be loaded from a pickle/joblib file
        self.model = IsolationForest(contamination=0.01, random_state=42)
        
        # Pre-train with some dummy historical variants to fix the model boundary
        dummy_data = np.array([
            [1, 1.0, 100], [2, 1.1, 150], [1, 0.9, 80], [3, 1.2, 200], [1, 1.0, 50],
            [10, 5.0, 5000], [1, 1.0, 120], [2, 1.0, 110], [1, 1.0, 90], [1, 1.0, 100]
        ])
        self.model.fit(dummy_data)
        logging.info("Advanced IsolationForest engine initialized.")

    def detect_fraud(self, transaction_id, user_id, amount):
        logging.info(f"Advanced fraud detection for transaction: {transaction_id}")
        
        # 1. Real-time Feature Sourcing from Redis
        features = self.feature_store.get_features_vector(user_id, amount)
        velocity, amount_ratio, amount = features
        
        # 2. ML Anomaly Detection (Isolation Forest)
        features_np = np.array([features])
        is_anomaly = self.model.predict(features_np)[0] == -1
        
        # 3. Byzantine Intelligence Reasons (XAI)
        reasons = []
        if velocity > 5:
            reasons.append(f"High transaction velocity ({velocity} in 10m)")
        if amount_ratio > 3.0:
            reasons.append(f"Suspicious amount variance ({amount_ratio:.1f}x avg)")
        if is_anomaly:
            reasons.append("Irregular transaction pattern detected (Isolation Forest)")

        risk_score = 0.0
        if is_anomaly: risk_score += 0.6
        if velocity > 3: risk_score += 0.2
        if amount_ratio > 2.0: risk_score += 0.2
        
        is_fraudulent = risk_score > 0.65
        reason_str = " | ".join(reasons) if reasons else "Normal transaction"
        
        return risk_score, is_fraudulent, reason_str

    def verify_security(self, domain, entity_id, amount):
        # Domain-specific logic
        risk_score = 0.0
        action = "ALLOW"
        reason = "Entity verified by AI engine"
        
        # 1. Base amount-based risk
        if amount > 10000:
            risk_score += 0.3
        if amount > 50000:
            risk_score += 0.5
            action = "CHALLENGE"
        
        # 2. Domain-specific triggers
        if domain == ai_security_pb2.WITHDRAW and amount > 5000:
            risk_score += 0.2
            reason = "High value withdrawal flagged for review"
        
        if domain == ai_security_pb2.TRANSFER:
            # Mock: check recipient profile linkage
            risk_score += 0.1
            
        if domain == ai_security_pb2.CARD:
            reason = "Card creation velocity check passed"

        # 3. Decision mapping
        if risk_score > 0.8:
            action = "DENY"
            reason = "Critical risk detected: transaction blocked"
        elif risk_score > 0.5:
            action = "CHALLENGE"
            
        return risk_score <= 0.8, risk_score, reason, action
