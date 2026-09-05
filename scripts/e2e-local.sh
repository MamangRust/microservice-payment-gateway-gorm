#!/bin/bash
# run-e2e.sh — One-shot: launch all services, wait for health, run hurl tests
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
BIN_DIR="/tmp/e2e-bin"
LOG_DIR="/tmp/e2e-logs"
COMPOSE_FILE="deployments/local/docker-compose.infra.yml"
mkdir -p "$LOG_DIR"

RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'
info()  { echo -e "${BLUE}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]${NC} $*"; }
fail()  { echo -e "${RED}[FAIL]${NC} $*"; }

cat > /tmp/e2e-common.env <<'ENVEOF'
APP_ENV=test
DB_DRIVER=postgres
DB_USERNAME=DRAGON
DB_PASSWORD=DRAGON
SECRET_KEY=yantopedia
KAFKA_BROKERS=localhost:9092
REDIS_ADDRS=localhost:6379
REDIS_PASSWORD=dragon_knight
REDIS_DB=0
GRPC_AUTH_ADDR=localhost:50051
GRPC_ROLE_ADDR=localhost:50052
GRPC_CARD_ADDR=localhost:50053
GRPC_MERCHANT_ADDR=localhost:50054
GRPC_USER_ADDR=localhost:50055
GRPC_SALDO_ADDR=localhost:50056
GRPC_TOPUP_ADDR=localhost:50057
GRPC_TRANSACTION_ADDR=localhost:50058
GRPC_TRANSFER_ADDR=localhost:50059
GRPC_WITHDRAW_ADDR=localhost:50060
GRPC_AI_SECURITY_ADDR=localhost:50051
ENVEOF

# ─── Seed roles ────────────────────────────────────────────────────────
info "Seeding roles ..."
docker compose -f "$COMPOSE_FILE" exec -T role-db psql -U DRAGON -d role_db -c "
INSERT INTO roles (role_name, created_at, updated_at) VALUES
  ('ROLE_ADMIN',now(),now()),('ROLE_USER',now(),now()),('ROLE_MERCHANT',now(),now()),
  ('ROLE_CUSTOMER',now(),now()),('ROLE_MODERATOR',now(),now()),('ROLE_SUPERVISOR',now(),now()),
  ('ROLE_ACCOUNTANT',now(),now()),('ROLE_SUPPORT',now(),now()),('ROLE_DEVELOPER',now(),now()),
  ('ROLE_MANAGER',now(),now())
ON CONFLICT (role_name) DO NOTHING;" 2>&1 | tail -1
ok "Roles seeded"

docker compose -f "$COMPOSE_FILE" exec -T redis redis-cli -a dragon_knight FLUSHALL 2>/dev/null | tail -1

# ─── Launch all services ───────────────────────────────────────────────
info "Launching all services ..."
pkill -9 -f '/tmp/e2e-bin/' 2>/dev/null || true
sleep 2

# Use setsid so processes survive parent shell exit
launch() {
  local svc=$1; shift
  local extra="$@"
  setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/$svc' && exec env $extra '$BIN_DIR/$svc' > '$LOG_DIR/$svc.log' 2>&1" &
  info "  launched $svc"
}

launch auth     DB_HOST=localhost DB_PORT=5433 DB_NAME=auth_db
launch user     DB_HOST=localhost DB_PORT=5434 DB_NAME=user_db
launch role     DB_HOST=localhost DB_PORT=5435 DB_NAME=role_db
launch card     DB_HOST=localhost DB_PORT=5436 DB_NAME=card_db BILLING_CYCLE_DAY=1
launch merchant DB_HOST=localhost DB_PORT=5437 DB_NAME=merchant_db
launch saldo    DB_HOST=localhost DB_PORT=5438 DB_NAME=saldo_db
launch topup    DB_HOST=localhost DB_PORT=5439 DB_NAME=topup_db
launch transaction DB_HOST=localhost DB_PORT=5440 DB_NAME=transaction_db
launch transfer DB_HOST=localhost DB_PORT=5441 DB_NAME=transfer_db
launch withdraw DB_HOST=localhost DB_PORT=5442 DB_NAME=withdraw_db WITHDRAW_DAILY_LIMIT=10000000

# stats-reader
setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/stats-reader' && exec env CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight '$BIN_DIR/stats-reader' > '$LOG_DIR/stats-reader.log' 2>&1" &
info "  launched stats-reader"

# stats-writer
setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/stats-writer' && exec env CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight '$BIN_DIR/stats-writer' > '$LOG_DIR/stats-writer.log' 2>&1" &
info "  launched stats-writer"

# email
setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/email' && exec env SMTP_SERVER=smtp.ethereal.email SMTP_PORT=587 SMTP_USER=giovani.roberts@ethereal.email SMTP_PASS=hwwvTzhWP2wW1y733m '$BIN_DIR/email' > '$LOG_DIR/email.log' 2>&1" &
info "  launched email"

# apigateway .env
cat > "$ROOT/service/apigateway/.env" <<'APENV'
REDIS_ADDRS=localhost:6379
REDIS_PASSWORD=dragon_knight
REDIS_DB=0
KAFKA_BROKERS=localhost:9092
SECRET_KEY=yantopedia
APENV

setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/apigateway' && exec env APP_ENV=development GRPC_AUTH=localhost:50051 GRPC_ROLE=localhost:50052 GRPC_CARD=localhost:50053 GRPC_MERCHANT=localhost:50054 GRPC_USER=localhost:50055 GRPC_SALDO=localhost:50056 GRPC_TOPUP=localhost:50057 GRPC_TRANSACTION=localhost:50058 GRPC_TRANSFER=localhost:50059 GRPC_WITHDRAW=localhost:50060 GRPC_STATS_READER=localhost:50062 GRPC_AI_SECURITY=localhost:50051 '$BIN_DIR/apigateway' > '$LOG_DIR/apigateway.log' 2>&1" &
info "  launched apigateway"

sleep 3
PROC_COUNT=$(ps aux | grep '/tmp/e2e-bin/' | grep -v grep | wc -l)
info "Running processes: $PROC_COUNT"

# ─── Wait for health ────────────────────────────────────────────────────
info "Waiting for all services to become healthy ..."
GW_URL="http://localhost:8080/api/auth/hello"
PORT_REGEX=':5005[1-9]|:50060|:50062'

svc_for_port() {
  case "$1" in
    50051) echo auth;;        50052) echo role;;
    50053) echo card;;        50054) echo merchant;;
    50055) echo user;;        50056) echo saldo;;
    50057) echo topup;;       50058) echo transaction;;
    50059) echo transfer;;    50060) echo withdraw;;
    50062) echo stats-reader;;
  esac
}

ok=0
for round in 1 2 3; do
  info "health check round $round ..."
  for i in $(seq 1 40); do
    gw=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 "$GW_URL" 2>/dev/null || echo "000")
    ports=$(ss -tln 2>/dev/null | grep -cE "$PORT_REGEX" || echo "0")
    if [ "$gw" = "200" ] && [ "$ports" -ge 11 ]; then ok=1; break; fi
    sleep 3
  done
  info "  round $round: apigateway=$gw gRPC_ports=$ports"
  [ "$ok" = "1" ] && break

  # restart missing gRPC services
  for port in 50051 50052 50053 50054 50055 50056 50057 50058 50059 50060 50062; do
    if ! ss -tln 2>/dev/null | grep -q ":$port "; then
      svc=$(svc_for_port "$port")
      info "  restarting $svc (port $port missing)"
      case "$svc" in
        auth)     setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/auth' && exec env DB_HOST=localhost DB_PORT=5433 DB_NAME=auth_db '$BIN_DIR/auth' > '$LOG_DIR/auth.log' 2>&1" &;;
        role)     setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/role' && exec env DB_HOST=localhost DB_PORT=5435 DB_NAME=role_db '$BIN_DIR/role' > '$LOG_DIR/role.log' 2>&1" &;;
        card)     setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/card' && exec env DB_HOST=localhost DB_PORT=5436 DB_NAME=card_db BILLING_CYCLE_DAY=1 '$BIN_DIR/card' > '$LOG_DIR/card.log' 2>&1" &;;
        merchant) setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/merchant' && exec env DB_HOST=localhost DB_PORT=5437 DB_NAME=merchant_db '$BIN_DIR/merchant' > '$LOG_DIR/merchant.log' 2>&1" &;;
        user)     setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/user' && exec env DB_HOST=localhost DB_PORT=5434 DB_NAME=user_db '$BIN_DIR/user' > '$LOG_DIR/user.log' 2>&1" &;;
        saldo)    setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/saldo' && exec env DB_HOST=localhost DB_PORT=5438 DB_NAME=saldo_db '$BIN_DIR/saldo' > '$LOG_DIR/saldo.log' 2>&1" &;;
        topup)    setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/topup' && exec env DB_HOST=localhost DB_PORT=5439 DB_NAME=topup_db '$BIN_DIR/topup' > '$LOG_DIR/topup.log' 2>&1" &;;
        transaction) setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/transaction' && exec env DB_HOST=localhost DB_PORT=5440 DB_NAME=transaction_db '$BIN_DIR/transaction' > '$LOG_DIR/transaction.log' 2>&1" &;;
        transfer) setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/transfer' && exec env DB_HOST=localhost DB_PORT=5441 DB_NAME=transfer_db '$BIN_DIR/transfer' > '$LOG_DIR/transfer.log' 2>&1" &;;
        withdraw) setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/withdraw' && exec env DB_HOST=localhost DB_PORT=5442 DB_NAME=withdraw_db WITHDRAW_DAILY_LIMIT=10000000 '$BIN_DIR/withdraw' > '$LOG_DIR/withdraw.log' 2>&1" &;;
        stats-reader) setsid bash -c "set -a; source /tmp/e2e-common.env; set +a; cd '$ROOT/service/stats-reader' && exec env CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight '$BIN_DIR/stats-reader' > '$LOG_DIR/stats-reader.log' 2>&1" &;;
      esac
      sleep 2
    fi
  done
done

# Re-seed roles + flush redis
info "Re-seeding roles + flushing Redis ..."
docker compose -f "$COMPOSE_FILE" exec -T role-db psql -U DRAGON -d role_db -c "
INSERT INTO roles (role_name, created_at, updated_at) VALUES
  ('ROLE_ADMIN',now(),now()),('ROLE_USER',now(),now()),('ROLE_MERCHANT',now(),now()),
  ('ROLE_CUSTOMER',now(),now()),('ROLE_MODERATOR',now(),now()),('ROLE_SUPERVISOR',now(),now()),
  ('ROLE_ACCOUNTANT',now(),now()),('ROLE_SUPPORT',now(),now()),('ROLE_DEVELOPER',now(),now()),
  ('ROLE_MANAGER',now(),now())
ON CONFLICT (role_name) DO NOTHING;" 2>&1 | tail -1
docker compose -f "$COMPOSE_FILE" exec -T redis redis-cli -a dragon_knight FLUSHALL 2>/dev/null | tail -1

if [ "$ok" != "1" ]; then
  fail "Services not healthy. Logs:"
  for f in "$LOG_DIR"/*.log; do echo "-- $(basename $f)"; tail -5 "$f"; done
  pkill -9 -f '/tmp/e2e-bin/' 2>/dev/null || true
  exit 1
fi

sleep 5
ok "All services healthy"

# ─── Run hurl tests ─────────────────────────────────────────────────────
info "Running hurl e2e tests ..."
rm -rf /tmp/hurl-e2e && cp -r "$ROOT/hurl" /tmp/hurl-e2e
sed -i 's/localhost:5000/localhost:8080/g' /tmp/hurl-e2e/*.hurl /tmp/hurl-e2e/auth/*.hurl 2>/dev/null || true

cd /tmp/hurl-e2e
PASS=0; FAIL=0

# Unique suffix shared across all hurl files in this run
SUFFIX=$(date +%s%N)
run_hurl() {
  local file=$1; shift
  if hurl --variable "year=$(date +%Y)" --variable "unique_suffix=$SUFFIX" "$@" "$file" 2>&1; then
    ok "  ✓ $file"; PASS=$((PASS+1))
  else
    fail "  ✗ $file"; FAIL=$((FAIL+1))
  fi
}

echo "────────────────────────────────────────"
echo "  Auth Tests"
echo "────────────────────────────────────────"
# Run register first, then auth (login+me+refresh), then login
run_hurl auth/register.hurl
run_hurl auth/auth.hurl
run_hurl auth/login.hurl

# Extract the auth_token captured by the auth test above
# (auth.hurl captures it last, but register.hurl also captures it)
# Re-login to get a fresh token for endpoint tests
info "Obtaining auth token ..."
TOKEN_RESP=$(curl -s --max-time 10 -X POST "http://localhost:8080/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"john.${SUFFIX}@hellodota.com\",\"password\":\"password123\"}" || true)

AUTH_TOKEN=""
if command -v python3 >/dev/null 2>&1; then
  AUTH_TOKEN=$(echo "$TOKEN_RESP" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("data",{}).get("access_token",""))' 2>/dev/null || echo "")
fi
if [ -z "$AUTH_TOKEN" ]; then
  AUTH_TOKEN=$(echo "$TOKEN_RESP" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
fi
[ -n "$AUTH_TOKEN" ] && ok "Auth token obtained" || fail "Could not get auth token"
export AUTH_TOKEN

echo "────────────────────────────────────────"
echo "  Endpoint Tests"
echo "────────────────────────────────────────"
for f in *.hurl; do
  [ -f "$f" ] || continue
  run_hurl "$f" --variable "auth_token=$AUTH_TOKEN"
done

cd "$ROOT"

# ─── Verify ClickHouse stats ───────────────────────────────────────────
info "Waiting for stats flush + verifying ClickHouse ..."
sleep 10
echo "--- ClickHouse row counts ---"
docker compose -f "$COMPOSE_FILE" exec -T clickhouse clickhouse-client \
  --user dragon --password dragon_knight -q "
SELECT 'card_events', count() FROM card_events
UNION ALL SELECT 'saldo_events', count() FROM saldo_events
UNION ALL SELECT 'topup_events', count() FROM topup_events
UNION ALL SELECT 'transaction_events', count() FROM transaction_events
UNION ALL SELECT 'transfer_events', count() FROM transfer_events
UNION ALL SELECT 'withdraw_events', count() FROM withdraw_events
UNION ALL SELECT 'merchant_events', count() FROM merchant_events" 2>/dev/null || warn "ClickHouse query failed"

echo ""
echo "========================================"
echo "  E2E Test Summary"
echo "========================================"
ok "Passed: $PASS"
if [ "$FAIL" -gt 0 ]; then
  fail "Failed: $FAIL"
  exit 1
else
  ok "All tests passed! 🎉"
fi

pkill -9 -f '/tmp/e2e-bin/' 2>/dev/null || true
rm -f "$ROOT/service/apigateway/.env"
