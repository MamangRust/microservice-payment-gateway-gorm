#!/bin/bash
# Launch all payment-gateway Go services locally (infra runs in docker).
set -u
ROOT=/home/hoover/monolith-payment-gateway-grpc
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
mkdir -p /tmp/e2e-logs

COMMON_GRPC="GRPC_AUTH_ADDR=localhost:50051 GRPC_ROLE_ADDR=localhost:50052 GRPC_CARD_ADDR=localhost:50053 GRPC_MERCHANT_ADDR=localhost:50054 GRPC_USER_ADDR=localhost:50055 GRPC_SALDO_ADDR=localhost:50056 GRPC_TOPUP_ADDR=localhost:50057 GRPC_TRANSACTION_ADDR=localhost:50058 GRPC_TRANSFER_ADDR=localhost:50059 GRPC_WITHDRAW_ADDR=localhost:50060 GRPC_AI_SECURITY_ADDR=localhost:50051"
COMMON_REDIS="REDIS_ADDRS=localhost:6379 REDIS_PASSWORD=dragon_knight REDIS_DB=0"
# One database per bounded context, reached through the infra PgBouncers
# (6432/6433/6434). Each service picks its own DB_<CONTEXT>_* set via the
# cluster prefix in its cmd/main.go; there is no per-service database anymore.
COMMON_DB="DB_DRIVER=postgres DB_USERNAME=postgres DB_PASSWORD=password DB_IDENTITY_HOST=localhost DB_IDENTITY_PORT=6432 DB_IDENTITY_NAME=pg_identity DB_PAYMENT_HOST=localhost DB_PAYMENT_PORT=6433 DB_PAYMENT_NAME=pg_payment DB_FINANCIAL_HOST=localhost DB_FINANCIAL_PORT=6434 DB_FINANCIAL_NAME=pg_financial"
COMMON="APP_ENV=test SECRET_KEY=yantopedia KAFKA_BROKERS=localhost:9092 $COMMON_DB $COMMON_REDIS $COMMON_GRPC"

launch() {
  local svc=$1; shift
  local dir=service/$svc
  (cd "$dir" && nohup env $COMMON "$@" /tmp/e2e-bin/$svc > /tmp/e2e-logs/$svc.log 2>&1 &)
  echo "launched $svc"
}

launch auth
launch user
launch role
launch card    BILLING_CYCLE_DAY=1
launch merchant
launch saldo
launch topup
launch transaction
launch transfer
launch withdraw WITHDRAW_DAILY_LIMIT=10000000

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
