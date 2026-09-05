#!/bin/bash
# Reproduce transaction.hurl failure: launch services, do the setup, inspect DB,
# then attempt the transaction create.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
mkdir -p /tmp/e2e-logs
B=http://localhost:8080

pkill -9 -f /tmp/e2e-bin 2>/dev/null; sleep 3

docker exec redis-local redis-cli -a dragon_knight FLUSHALL >/dev/null 2>&1

COMMON_GRPC="GRPC_AUTH_ADDR=localhost:50051 GRPC_ROLE_ADDR=localhost:50052 GRPC_CARD_ADDR=localhost:50053 GRPC_MERCHANT_ADDR=localhost:50054 GRPC_USER_ADDR=localhost:50055 GRPC_SALDO_ADDR=localhost:50056 GRPC_TOPUP_ADDR=localhost:50057 GRPC_TRANSACTION_ADDR=localhost:50058 GRPC_TRANSFER_ADDR=localhost:50059 GRPC_WITHDRAW_ADDR=localhost:50060 GRPC_AI_SECURITY_ADDR=localhost:50051"
COMMON="APP_ENV=test DB_DRIVER=postgres DB_USERNAME=DRAGON DB_PASSWORD=DRAGON SECRET_KEY=yantopedia KAFKA_BROKERS=localhost:9092 REDIS_ADDRS=localhost:6379 REDIS_PASSWORD=dragon_knight REDIS_DB=0 $COMMON_GRPC"

launch() { local svc=$1; shift; (cd service/$svc && env $COMMON "$@" /tmp/e2e-bin/$svc > /tmp/e2e-logs/$svc.log 2>&1 &); }

launch auth DB_HOST=localhost DB_PORT=5433 DB_NAME=auth_db
launch user DB_HOST=localhost DB_PORT=5434 DB_NAME=user_db
launch role DB_HOST=localhost DB_PORT=5435 DB_NAME=role_db
launch card DB_HOST=localhost DB_PORT=5436 DB_NAME=card_db BILLING_CYCLE_DAY=1
launch merchant DB_HOST=localhost DB_PORT=5437 DB_NAME=merchant_db
launch saldo DB_HOST=localhost DB_PORT=5438 DB_NAME=saldo_db
launch topup DB_HOST=localhost DB_PORT=5439 DB_NAME=topup_db
launch transaction DB_HOST=localhost DB_PORT=5440 DB_NAME=transaction_db
launch transfer DB_HOST=localhost DB_PORT=5441 DB_NAME=transfer_db
launch withdraw DB_HOST=localhost DB_PORT=5442 DB_NAME=withdraw_db WITHDRAW_DAILY_LIMIT=10000000
(cd service/stats-reader && env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight /tmp/e2e-bin/stats-reader > /tmp/e2e-logs/stats-reader.log 2>&1 &)
(cd service/stats-writer && env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight KAFKA_BROKERS=localhost:9092 /tmp/e2e-bin/stats-writer > /tmp/e2e-logs/stats-writer.log 2>&1 &)
cat > service/apigateway/.env <<'ENV'
REDIS_ADDRS=localhost:6379
REDIS_PASSWORD=dragon_knight
REDIS_DB=0
KAFKA_BROKERS=localhost:9092
SECRET_KEY=yantopedia
ENV
(cd service/apigateway && env APP_ENV=development GRPC_AUTH=localhost:50051 GRPC_ROLE=localhost:50052 GRPC_CARD=localhost:50053 GRPC_MERCHANT=localhost:50054 GRPC_USER=localhost:50055 GRPC_SALDO=localhost:50056 GRPC_TOPUP=localhost:50057 GRPC_TRANSACTION=localhost:50058 GRPC_TRANSFER=localhost:50059 GRPC_WITHDRAW=localhost:50060 GRPC_STATS_READER=localhost:50062 GRPC_AI_SECURITY=localhost:50051 /tmp/e2e-bin/apigateway > /tmp/e2e-logs/apigateway.log 2>&1 &)

echo "waiting for services..."
ok=0
for i in $(seq 1 40); do
  gw=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 "$B/api/auth/hello" 2>/dev/null)
  ports=$(ss -tln 2>/dev/null | grep -cE ':5005[1-9]|:50060|:50062')
  [ "$gw" = "200" ] && [ "$ports" -ge 11 ] && { ok=1; break; }
  sleep 3
done
echo "gw=$gw ports=$ports"
[ "$ok" = "1" ] || { echo "services unhealthy"; exit 1; }
sleep 5

# login (user exists from prior runs) or register
TOKEN=$(curl -s -X POST "$B/api/auth/login" -H "Content-Type: application/json" -d '{"email":"john.doe@hellodota.com","password":"password123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["access_token"])')
echo "token acquired: ${TOKEN:0:20}..."
AUTH="Authorization: Bearer $TOKEN"

USER_ID=$(curl -s "$B/api/auth/me" -H "$AUTH" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')
echo "user_id=$USER_ID"

# 1. customer card
CARD=$(curl -s -X POST "$B/api/card-command/create" -H "$AUTH" -H "Content-Type: application/json" -d "{\"user_id\": $USER_ID, \"card_type\": \"debit\", \"expire_date\": \"2030-12-31T00:00:00Z\", \"cvv\": \"123\", \"card_provider\": \"visa\"}")
CARD_NUM=$(echo "$CARD" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["card_number"])')
echo "customer card=$CARD_NUM"

# 2. saldo + adjustment
curl -s -X POST "$B/api/saldo-command/create" -H "$AUTH" -H "Content-Type: application/json" -d "{\"card_number\": \"$CARD_NUM\", \"total_balance\": 1}" > /dev/null
curl -s -X POST "$B/api/saldo-command/adjustment" -H "$AUTH" -H "Content-Type: application/json" -d "{\"card_number\": \"$CARD_NUM\", \"delta\": 999999, \"operation_id\": \"repro-seed-$CARD_NUM\", \"source_type\": \"test_seed\", \"source_id\": \"repro\"}" > /dev/null

# 3. merchant
MERCH=$(curl -s -X POST "$B/api/merchant-command/create" -H "$AUTH" -H "Content-Type: application/json" -d "{\"name\": \"Repro Merchant\", \"user_id\": $USER_ID}")
MERCH_ID=$(echo "$MERCH" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')
MERCH_KEY=$(echo "$MERCH" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["api_key"])')
echo "merchant id=$MERCH_ID key=${MERCH_KEY:0:12}..."

# 4. merchant receiving card
MCARD=$(curl -s -X POST "$B/api/card-command/create" -H "$AUTH" -H "Content-Type: application/json" -d "{\"user_id\": $USER_ID, \"card_type\": \"credit\", \"expire_date\": \"2031-12-31T00:00:00Z\", \"cvv\": \"456\", \"card_provider\": \"mastercard\"}")
MCARD_NUM=$(echo "$MCARD" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["card_number"])')
echo "merchant card=$MCARD_NUM"

# 5. merchant saldo
curl -s -X POST "$B/api/saldo-command/create" -H "$AUTH" -H "Content-Type: application/json" -d "{\"card_number\": \"$MCARD_NUM\", \"total_balance\": 1000000}" > /dev/null

echo "=== CARD DB STATE (user $USER_ID) ==="
docker exec pg-local-card-db-1 psql -U DRAGON -d card_db -t -c "SELECT card_id, card_number, card_type, deleted_at FROM cards WHERE user_id = $USER_ID AND deleted_at IS NULL ORDER BY card_id LIMIT 5;"

echo "=== GET CARDBYUSERID RESULT (what FindCardByUserId sees) ==="
docker exec pg-local-card-db-1 psql -U DRAGON -d card_db -t -c "SELECT card_id, card_number FROM cards WHERE user_id = $USER_ID AND deleted_at IS NULL LIMIT 1;"

echo "=== TRANSACTION CREATE ==="
RESP=$(curl -s -w "\nHTTP:%{http_code}" -X POST "$B/api/transaction-command/create" \
  -H "$AUTH" -H "X-Api-Key: $MERCH_KEY" -H "Idempotency-Key: repro-$CARD_NUM" -H "Content-Type: application/json" \
  -d "{\"card_number\": \"$CARD_NUM\", \"amount\": 50000, \"payment_method\": \"visa\", \"merchant_id\": $MERCH_ID, \"transaction_time\": \"2026-08-01T10:00:00Z\"}")
echo "$RESP" | tail -3

echo "=== SALDO LOG (last credit/debit) ==="
grep -E "CreditSaldo|DebitSaldo|FindByCardNumber" /tmp/e2e-logs/saldo.log | tail -6
echo "=== CARD LOG (FindByUserIdCard) ==="
grep -E "FindByUserIdCard" /tmp/e2e-logs/card.log | tail -3
echo "=== TRANSACTION LOG ==="
grep -E "request failed|CreateTransaction" /tmp/e2e-logs/transaction.log | tail -4
echo "=== DONE ==="
