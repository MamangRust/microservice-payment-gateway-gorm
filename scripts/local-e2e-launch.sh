#!/bin/bash
# Launch all payment-gateway Go services locally (infra runs in docker).
set -u
ROOT=/home/hoover/monolith-payment-gateway-grpc
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
mkdir -p /tmp/e2e-logs

COMMON_GRPC="GRPC_AUTH_ADDR=localhost:50051 GRPC_ROLE_ADDR=localhost:50052 GRPC_CARD_ADDR=localhost:50053 GRPC_MERCHANT_ADDR=localhost:50054 GRPC_USER_ADDR=localhost:50055 GRPC_SALDO_ADDR=localhost:50056 GRPC_TOPUP_ADDR=localhost:50057 GRPC_TRANSACTION_ADDR=localhost:50058 GRPC_TRANSFER_ADDR=localhost:50059 GRPC_WITHDRAW_ADDR=localhost:50060 GRPC_AI_SECURITY_ADDR=localhost:50051"
COMMON_REDIS="REDIS_ADDRS=localhost:6379 REDIS_PASSWORD=dragon_knight REDIS_DB=0"
COMMON="APP_ENV=test SECRET_KEY=yantopedia KAFKA_BROKERS=localhost:9092 $COMMON_REDIS $COMMON_GRPC"

launch() {
  local svc=$1; shift
  local dir=service/$svc
  (cd "$dir" && nohup env $COMMON "$@" /tmp/e2e-bin/$svc > /tmp/e2e-logs/$svc.log 2>&1 &)
  echo "launched $svc"
}

# DB port per service: auth 5433, user 5434, role 5435, card 5436,
# merchant 5437, saldo 5438, topup 5439, transaction 5440, transfer 5441, withdraw 5442
launch auth    DB_HOST=localhost DB_PORT=5433 DB_NAME=auth_db
launch user    DB_HOST=localhost DB_PORT=5434 DB_NAME=user_db
launch role    DB_HOST=localhost DB_PORT=5435 DB_NAME=role_db
launch card    DB_HOST=localhost DB_PORT=5436 DB_NAME=card_db BILLING_CYCLE_DAY=1
launch merchant DB_HOST=localhost DB_PORT=5437 DB_NAME=merchant_db
launch saldo   DB_HOST=localhost DB_PORT=5438 DB_NAME=saldo_db
launch topup   DB_HOST=localhost DB_PORT=5439 DB_NAME=topup_db
launch transaction DB_HOST=localhost DB_PORT=5440 DB_NAME=transaction_db
launch transfer DB_HOST=localhost DB_PORT=5441 DB_NAME=transfer_db
launch withdraw DB_HOST=localhost DB_PORT=5442 DB_NAME=withdraw_db WITHDRAW_DAILY_LIMIT=10000000

# stats services (ClickHouse + Kafka)
(cd service/stats-reader && nohup env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight /tmp/e2e-bin/stats-reader > /tmp/e2e-logs/stats-reader.log 2>&1 &)
echo "launched stats-reader"
(cd service/stats-writer && nohup env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight KAFKA_BROKERS=localhost:9092 /tmp/e2e-bin/stats-writer > /tmp/e2e-logs/stats-writer.log 2>&1 &)
echo "launched stats-writer"

# email service (kafka consumer + SMTP)
(cd service/email && nohup env APP_ENV=test KAFKA_BROKERS=localhost:9092 SMTP_SERVER=smtp.ethereal.email SMTP_PORT=587 SMTP_USER=giovani.roberts@ethereal.email SMTP_PASS=hwwvTzhWP2wW1y733m /tmp/e2e-bin/email > /tmp/e2e-logs/email.log 2>&1 &)
echo "launched email"

# apigateway: reads a config file for post-prefix keys -> generate .env in its dir
cat > service/apigateway/.env <<'ENV'
REDIS_ADDRS=localhost:6379
REDIS_PASSWORD=dragon_knight
REDIS_DB=0
KAFKA_BROKERS=localhost:9092
SECRET_KEY=yantopedia
ENV
(cd service/apigateway && nohup env APP_ENV=development GRPC_AUTH=localhost:50051 GRPC_ROLE=localhost:50052 GRPC_CARD=localhost:50053 GRPC_MERCHANT=localhost:50054 GRPC_USER=localhost:50055 GRPC_SALDO=localhost:50056 GRPC_TOPUP=localhost:50057 GRPC_TRANSACTION=localhost:50058 GRPC_TRANSFER=localhost:50059 GRPC_WITHDRAW=localhost:50060 GRPC_STATS_READER=localhost:50062 GRPC_AI_SECURITY=localhost:50051 /tmp/e2e-bin/apigateway > /tmp/e2e-logs/apigateway.log 2>&1 &)
echo "launched apigateway"

echo "ALL LAUNCHED"
