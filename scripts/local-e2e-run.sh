#!/bin/bash
# Local e2e: Go services on host + infra in docker. Runs inside ONE command
# because backgrounded processes are killed when the tool call returns.
# Services whose gRPC bind races (EADDRINUSE seen intermittently in batch
# launches) are auto-restarted by the health loop.
set -u
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
mkdir -p /tmp/e2e-logs

echo "===== [1/7] cleanup old processes"
pkill -9 -f /tmp/e2e-bin 2>/dev/null; sleep 3
if ss -tlnp 2>/dev/null | grep -E ':5005[1-9]|:50060|:50062|:8080'; then
  echo "!!! ports occupied pre-flight (holders below):"
  ss -tlnp 2>/dev/null | grep -E ':5005[1-9]|:50060|:50062|:8080'
else
  echo "pre-flight ports free"
fi

echo "===== [2/7] seeding base roles (role_db)"
# Register assigns ROLE_ADMIN to new users, so the role must pre-exist.
docker exec pg-local-role-db-1 psql -U DRAGON -d role_db -c "
INSERT INTO roles (role_name, created_at, updated_at) VALUES
  ('ROLE_ADMIN',now(),now()),('ROLE_USER',now(),now()),('ROLE_MERCHANT',now(),now()),
  ('ROLE_CUSTOMER',now(),now()),('ROLE_MODERATOR',now(),now()),('ROLE_SUPERVISOR',now(),now()),
  ('ROLE_ACCOUNTANT',now(),now()),('ROLE_SUPPORT',now(),now()),('ROLE_DEVELOPER',now(),now()),
  ('ROLE_MANAGER',now(),now())
ON CONFLICT (role_name) DO NOTHING;" 2>&1 | tail -1

echo "===== [2b/7] flushing redis cache"
# redis-local persists across runs; stale cached responses (e.g. empty
# FindAllRole results cached before roles were seeded) break the e2e.
docker exec redis-local redis-cli -a dragon_knight FLUSHALL 2>&1 | tail -1

echo "===== [3/7] launching services"
COMMON_GRPC="GRPC_AUTH_ADDR=localhost:50051 GRPC_ROLE_ADDR=localhost:50052 GRPC_CARD_ADDR=localhost:50053 GRPC_MERCHANT_ADDR=localhost:50054 GRPC_USER_ADDR=localhost:50055 GRPC_SALDO_ADDR=localhost:50056 GRPC_TOPUP_ADDR=localhost:50057 GRPC_TRANSACTION_ADDR=localhost:50058 GRPC_TRANSFER_ADDR=localhost:50059 GRPC_WITHDRAW_ADDR=localhost:50060 GRPC_AI_SECURITY_ADDR=localhost:50051"
COMMON="APP_ENV=test DB_DRIVER=postgres DB_USERNAME=DRAGON DB_PASSWORD=DRAGON SECRET_KEY=yantopedia KAFKA_BROKERS=localhost:9092 REDIS_ADDRS=localhost:6379 REDIS_PASSWORD=dragon_knight REDIS_DB=0 $COMMON_GRPC"

launch() {
  local svc=$1; shift
  pkill -9 -f "e2e-bin/$svc" 2>/dev/null
  sleep 1
  (cd service/$svc && env $COMMON "$@" /tmp/e2e-bin/$svc > /tmp/e2e-logs/$svc.log 2>&1 &)
  echo "  launched $svc"
}

launch_all() {
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

  (cd service/stats-reader && env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight /tmp/e2e-bin/stats-reader > /tmp/e2e-logs/stats-reader.log 2>&1 &)
  echo "  launched stats-reader"
  (cd service/stats-writer && env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight KAFKA_BROKERS=localhost:9092 /tmp/e2e-bin/stats-writer > /tmp/e2e-logs/stats-writer.log 2>&1 &)
  echo "  launched stats-writer"
  (cd service/email && env APP_ENV=test KAFKA_BROKERS=localhost:9092 SMTP_SERVER=smtp.ethereal.email SMTP_PORT=587 SMTP_USER=giovani.roberts@ethereal.email SMTP_PASS=hwwvTzhWP2wW1y733m /tmp/e2e-bin/email > /tmp/e2e-logs/email.log 2>&1 &)
  echo "  launched email"

  cat > service/apigateway/.env <<'ENV'
REDIS_ADDRS=localhost:6379
REDIS_PASSWORD=dragon_knight
REDIS_DB=0
KAFKA_BROKERS=localhost:9092
SECRET_KEY=yantopedia
ENV
  (cd service/apigateway && env APP_ENV=development GRPC_AUTH=localhost:50051 GRPC_ROLE=localhost:50052 GRPC_CARD=localhost:50053 GRPC_MERCHANT=localhost:50054 GRPC_USER=localhost:50055 GRPC_SALDO=localhost:50056 GRPC_TOPUP=localhost:50057 GRPC_TRANSACTION=localhost:50058 GRPC_TRANSFER=localhost:50059 GRPC_WITHDRAW=localhost:50060 GRPC_STATS_READER=localhost:50062 GRPC_AI_SECURITY=localhost:50051 /tmp/e2e-bin/apigateway > /tmp/e2e-logs/apigateway.log 2>&1 &)
  echo "  launched apigateway"
}

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

# Actually launch everything (the function is defined above, call it now)
launch_all

echo "===== [4/7] waiting for health (with auto-restart)"
# NOTE: /health is protected by the auth middleware (401) - use the public
# /api/auth/hello endpoint for readiness, like run_tests.sh does.
GW_URL="http://localhost:8080/api/auth/hello"
PORT_REGEX=':5005[1-9]|:50060|:50062'

ok=0
for round in 1 2 3; do
  for i in $(seq 1 40); do
    gw=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 "$GW_URL" 2>/dev/null)
    ports=$(ss -tln 2>/dev/null | grep -cE "$PORT_REGEX")
    if [ "$gw" = "200" ] && [ "$ports" -ge 11 ]; then ok=1; break; fi
    sleep 3
  done
  echo "round $round: apigateway=$gw gRPC_ports=$ports"
  [ "$ok" = "1" ] && break
  # restart missing gRPC services (bind race) and retry
  for port in 50051 50052 50053 50054 50055 50056 50057 50058 50059 50060 50062; do
    if ! ss -tln 2>/dev/null | grep -q ":$port "; then
      svc=$(svc_for_port "$port")
      echo "  restarting $svc (port $port missing)"
      case "$svc" in
        auth)         launch auth         DB_HOST=localhost DB_PORT=5433 DB_NAME=auth_db;;
        role)         launch role         DB_HOST=localhost DB_PORT=5435 DB_NAME=role_db;;
        card)         launch card         DB_HOST=localhost DB_PORT=5436 DB_NAME=card_db BILLING_CYCLE_DAY=1;;
        merchant)     launch merchant     DB_HOST=localhost DB_PORT=5437 DB_NAME=merchant_db;;
        user)         launch user         DB_HOST=localhost DB_PORT=5434 DB_NAME=user_db;;
        saldo)        launch saldo        DB_HOST=localhost DB_PORT=5438 DB_NAME=saldo_db;;
        topup)        launch topup        DB_HOST=localhost DB_PORT=5439 DB_NAME=topup_db;;
        transaction)  launch transaction  DB_HOST=localhost DB_PORT=5440 DB_NAME=transaction_db;;
        transfer)     launch transfer     DB_HOST=localhost DB_PORT=5441 DB_NAME=transfer_db;;
        withdraw)     launch withdraw     DB_HOST=localhost DB_PORT=5442 DB_NAME=withdraw_db WITHDRAW_DAILY_LIMIT=10000000;;
        stats-reader) (cd service/stats-reader && env APP_ENV=test CLICKHOUSE_ADDR=localhost:9000 CLICKHOUSE_DATABASE=default CLICKHOUSE_USERNAME=dragon CLICKHOUSE_PASSWORD=dragon_knight /tmp/e2e-bin/stats-reader > /tmp/e2e-logs/stats-reader.log 2>&1 &); echo "  launched stats-reader";;
      esac
    fi
  done
done

echo "===== [4b/7] re-seeding base roles + flushing redis (fresh-DB safe)"
# On a fresh DB the role service has only just migrated, so the earlier
# [2/7] seed hit a missing `roles` table. Seed again now that the table
# exists, then flush redis so no empty result cached before seeding leaks.
docker exec pg-local-role-db-1 psql -U DRAGON -d role_db -c "
INSERT INTO roles (role_name, created_at, updated_at) VALUES
  ('ROLE_ADMIN',now(),now()),('ROLE_USER',now(),now()),('ROLE_MERCHANT',now(),now()),
  ('ROLE_CUSTOMER',now(),now()),('ROLE_MODERATOR',now(),now()),('ROLE_SUPERVISOR',now(),now()),
  ('ROLE_ACCOUNTANT',now(),now()),('ROLE_SUPPORT',now(),now()),('ROLE_DEVELOPER',now(),now()),
  ('ROLE_MANAGER',now(),now())
ON CONFLICT (role_name) DO NOTHING;" 2>&1 | tail -1
docker exec redis-local redis-cli -a dragon_knight FLUSHALL 2>&1 | tail -1

if [ "$ok" != "1" ]; then
  echo "!!! services not healthy; current listeners:"
  ss -tlnp 2>/dev/null | grep -E "$PORT_REGEX|:8080" || echo "(no listeners)"
  echo "!!! logs with bind errors:"
  grep -l "address already in use" /tmp/e2e-logs/*.log 2>/dev/null
  echo "!!! recent log tails:"
  for f in /tmp/e2e-logs/*.log; do echo "-- $f"; tail -4 "$f"; done
  pkill -9 -f /tmp/e2e-bin 2>/dev/null
  exit 1
fi
sleep 8

echo "===== [5/7] preparing hurl (port 8080)"
rm -rf /tmp/hurl-e2e && cp -r hurl /tmp/hurl-e2e
sed -i 's/localhost:5000/localhost:8080/g' /tmp/hurl-e2e/*.hurl /tmp/hurl-e2e/run_tests.sh /tmp/hurl-e2e/auth/*.hurl

echo "===== [6/7] running hurl suite"
cd /tmp/hurl-e2e
{
  echo "########## auth/register.hurl"
  hurl --variable year=$(date +%Y) auth/register.hurl 2>&1 | tail -4
  echo "########## auth/auth.hurl"
  hurl --variable year=$(date +%Y) auth/auth.hurl 2>&1 | tail -4
  echo "########## auth/login.hurl"
  hurl --variable year=$(date +%Y) auth/login.hurl 2>&1 | tail -4
} > /tmp/e2e-auth.log 2>&1
bash run_tests.sh > /tmp/e2e-main.log 2>&1
echo "run_tests exit=$?"
cd "$ROOT"

echo "===== [7/7] verifying stats (clickhouse) + cleanup"
# stats-writer flushes every 5s; wait so the last batch (e.g. transaction
# events published at the very end of the hurl suite) lands before counting.
sleep 8
echo "--- clickhouse row counts ---"
curl -s "http://localhost:8123/?query=SELECT%20'card_events',count()%20FROM%20card_events%20UNION%20ALL%20SELECT%20'saldo_events',count()%20FROM%20saldo_events%20UNION%20ALL%20SELECT%20'topup_events',count()%20FROM%20topup_events%20UNION%20ALL%20SELECT%20'transaction_events',count()%20FROM%20transaction_events%20UNION%20ALL%20SELECT%20'transfer_events',count()%20FROM%20transfer_events%20UNION%20ALL%20SELECT%20'withdraw_events',count()%20FROM%20withdraw_events%20UNION%20ALL%20SELECT%20'merchant_events',count()%20FROM%20merchant_events" -u dragon:dragon_knight 2>&1
pkill -9 -f /tmp/e2e-bin 2>/dev/null
rm -f service/apigateway/.env
echo "=== DONE ==="
