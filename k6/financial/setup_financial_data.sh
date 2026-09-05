#!/bin/bash
# Setup test data for k6/financial/financial_concurrency.js
# Creates two cards + saldo records via the API Gateway, then prints
# the env vars to pass to k6.
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:5000}"
EMAIL="${EMAIL:-john.doe@hellodota.com}"
PASSWORD="${PASSWORD:-password123}"

login_resp=$(curl -fsS -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
TOKEN=$(echo "$login_resp" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("data",{}).get("access_token",""))')
if [ -z "$TOKEN" ]; then
  echo "LOGIN_FAILED: $login_resp" >&2
  exit 1
fi

ME=$(curl -fsS "$BASE_URL/api/auth/me" -H "Authorization: Bearer $TOKEN")
USER_ID=$(echo "$ME" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("data",{}).get("id",""))')
if [ -z "$USER_ID" ]; then
  echo "AUTH_ME_FAILED: invalid user payload" >&2
  exit 1
fi
echo "user_id=$USER_ID"

# ---- create card 1 (sender / topup / withdraw / transaction) ----
card1_resp=$(curl -fsS -X POST "$BASE_URL/api/card-command/create" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"user_id\": $USER_ID, \"card_type\": \"debit\", \"expire_date\": \"2030-12-31T00:00:00Z\", \"cvv\": \"123\", \"card_provider\": \"visa\"}")
CARD1=$(echo "$card1_resp" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("data",{}).get("card_number",""))')
if [ -z "$CARD1" ]; then
  echo "CARD1_CREATE_FAILED: invalid card payload" >&2
  exit 1
fi
echo "card1=$CARD1"

# ---- create card 2 (transfer receiver) ----
card2_resp=$(curl -fsS -X POST "$BASE_URL/api/card-command/create" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"user_id\": $USER_ID, \"card_type\": \"debit\", \"expire_date\": \"2030-12-31T00:00:00Z\", \"cvv\": \"123\", \"card_provider\": \"visa\"}")
CARD2=$(echo "$card2_resp" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("data",{}).get("card_number",""))')
if [ -z "$CARD2" ]; then
  echo "CARD2_CREATE_FAILED: invalid card payload" >&2
  exit 1
fi
echo "card2=$CARD2"

# ---- saldo for both cards ----
saldo1_resp=$(curl -fsS -X POST "$BASE_URL/api/saldo-command/create" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"card_number\": \"$CARD1\", \"total_balance\": 500000000}")
saldo2_resp=$(curl -fsS -X POST "$BASE_URL/api/saldo-command/create" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"card_number\": \"$CARD2\", \"total_balance\": 500000000}")
python3 - "$saldo1_resp" "$saldo2_resp" <<'PY'
import json, sys
for i, raw in enumerate(sys.argv[1:], 1):
    payload = json.loads(raw)
    if payload.get("status") != "success" or not payload.get("data", {}).get("card_number"):
        raise SystemExit(f"SALDO{i}_CREATE_FAILED: invalid saldo payload")
PY

# ---- merchant for transaction ----
MERCHANT_ID="${MERCHANT_ID:-1}"
API_KEY=$(curl -fsS "$BASE_URL/api/merchant-query" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; d=json.load(sys.stdin); data=d.get("data") or []; print(data[0].get("api_key","") if data else "")' 2>/dev/null || echo "")
if [ -z "$API_KEY" ]; then
  echo "MERCHANT_LOOKUP_FAILED: merchant API key is empty" >&2
  exit 1
fi
echo "merchant_api_key=<redacted>"

echo "---"
echo "RUN_K6: k6 run -e TOKEN=\"Bearer <paste-token>\" -e CARD_NUMBER=\"$CARD1\" -e TRANSFER_TO=\"$CARD2\" -e API_KEY=\"<paste-api-key>\" -e MERCHANT_ID=$MERCHANT_ID k6/financial/financial_concurrency.js"
