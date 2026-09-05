import logging
import json
import threading
from kafka import KafkaConsumer

class TransactionConsumer(threading.Thread):
    def __init__(self, bootstrap_servers=['localhost:9092']):
        super().__init__()
        self.bootstrap_servers = bootstrap_servers
        self.daemon = True
        self.running = True

    def run(self):
        logging.info(f"Connecting to Kafka at {self.bootstrap_servers}...")
        try:
            consumer = KafkaConsumer(
                'stats-topic-transaction-events',
                'stats-topic-card-events',
                'stats-topic-merchant-events',
                'stats-topic-saldo-events',
                'stats-topic-topup-events',
                'stats-topic-transfer-events',
                'stats-topic-withdraw-events',
                bootstrap_servers=self.bootstrap_servers,
                auto_offset_reset='earliest',
                enable_auto_commit=True,
                group_id='ai-security-group',
                value_deserializer=lambda x: json.loads(x.decode('utf-8'))
            )
            
            logging.info("Subscribed to all domain security topics.")

            for message in consumer:
                if not self.running:
                    break
                event = message.value
                logging.info(f"Received transaction event: {event.get('transaction_id')}")
                self.process_event(event)
        except Exception as e:
            logging.error(f"Kafka consumer error: {e}")

    def process_event(self, event):
        # In a real scenario, this would update user risk profiles
        # or store data for retraining
        logging.info(f"Processing event for ML learning: {event.get('transaction_id')}")

    def stop(self):
        self.running = False
