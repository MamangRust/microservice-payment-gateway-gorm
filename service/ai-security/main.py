import sys
import os
import logging
from concurrent import futures
import grpc

# Add the generated folder to the path for protobuf imports - MUST BE FIRST
sys.path.append(os.path.join(os.path.dirname(__file__), 'generated/ai_security'))

from ai_security import ai_security_pb2_grpc
from service import AISecurityService
from kafka_consumer import TransactionConsumer

def serve():
    # Start Kafka Consumer in a separate thread
    kafka_servers = os.environ.get('KAFKA_SERVERS', 'localhost:9092').split(',')
    consumer = TransactionConsumer(bootstrap_servers=kafka_servers)
    consumer.start()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ai_security_pb2_grpc.add_AISecurityServiceServicer_to_server(AISecurityService(), server)
    server.add_insecure_port('[::]:50051')
    logging.info("AI Security Service starting on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    serve()
