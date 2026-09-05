from redis.cluster import RedisCluster, ClusterNode
import json
import os
import time

class FeatureStore:
    def __init__(self, startup_nodes=None):
        if startup_nodes is None:
            # Fallback to single node if not provided
            host = os.getenv('REDIS_HOST', 'localhost')
            port = int(os.getenv('REDIS_PORT', 6379))
            startup_nodes = [ClusterNode(host, port)]
        else:
            # Ensure they are ClusterNode objects if passed as dicts
            normalized_nodes = []
            for node in startup_nodes:
                if isinstance(node, dict):
                    normalized_nodes.append(ClusterNode(node['host'], node['port']))
                else:
                    normalized_nodes.append(node)
            startup_nodes = normalized_nodes
        
        password = os.getenv('REDIS_PASSWORD', 'dragon_knight')
        
        self.redis = RedisCluster(
            startup_nodes=startup_nodes,
            password=password,
            decode_responses=True,
            skip_full_coverage_check=True # Useful for scaled clusters
        )
        self.velocity_window = 600 # 10 minutes

    def get_transaction_velocity(self, user_id):
        """Returns the number of transactions by this user in the last 10 minutes."""
        key = f"feature:velocity:user:{user_id}"
        now = time.time()
        
        # Add current transaction to the sliding window
        self.redis.zadd(key, {str(now): now})
        
        # Remove old transactions
        self.redis.zremrangebyscore(key, 0, now - self.velocity_window)
        
        # Get count
        velocity = self.redis.zcard(key)
        return velocity

    def get_amount_stats(self, user_id, current_amount):
        """Returns variance/ratio compared to user's average amount."""
        key = f"feature:stats:user:{user_id}"
        
        stats_json = self.redis.get(key)
        if not stats_json:
            # First transaction, initialize stats
            initial_stats = {"avg": current_amount, "count": 1}
            self.redis.set(key, json.dumps(initial_stats))
            return 1.0 # Ratio is 1.0 for first transaction
        
        stats = json.loads(stats_json)
        avg = stats.get('avg', current_amount)
        count = stats.get('count', 1)
        
        # Calculate ratio
        ratio = current_amount / avg if avg > 0 else 1.0
        
        # Update running average
        new_avg = (avg * count + current_amount) / (count + 1)
        self.redis.set(key, json.dumps({"avg": new_avg, "count": count + 1}))
        
        return ratio

    def get_features_vector(self, user_id, amount):
        """Compiles a numerical vector for the ML model."""
        velocity = self.get_transaction_velocity(user_id)
        amount_ratio = self.get_amount_stats(user_id, amount)
        
        return [velocity, amount_ratio, amount]
