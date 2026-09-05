// Financial mutation / error-contract test (Phase 5, P2 load-test mutation).
//
// Unlike financial_concurrency.js (which hammers happy-path creates), this
// suite deliberately sends retries, conflicts, and invalid payloads and asserts
// the error contract:
//
//   - retry with the SAME Idempotency-Key + payload -> replay, no double move
//   - same key + DIFFERENT payload -> 409 Conflict
//   - amount <= 0 -> 400 (P1.7 validation)
//   - amount > balance -> 409 Conflict (P1.5 error contract)
//   - transfer to the same card -> 400 (P1.7 validation)
//
// Configuration via environment variables (k6 -e KEY=VALUE):
//   BASE_URL     default http://localhost:5000
//   TOKEN        Bearer token (required)
//   API_KEY      merchant API key for transaction (optional)
//   CARD_NUMBER  default 1234567890123456
//   TRANSFER_TO  receiver card (default 6543210987654321)
//   MERCHANT_ID  merchant id for transaction (default 1)
//   TOPUP_METHOD valid method (default bca)
//
// Run: k6 run -e TOKEN='Bearer ...' -e CARD_NUMBER=... -e TRANSFER_TO=... k6/financial/financial_mutation.js

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:5000';
const TOKEN = __ENV.TOKEN || '';
const API_KEY = __ENV.API_KEY || 'merchant-key';
const CARD_NUMBER = __ENV.CARD_NUMBER || '1234567890123456';
const TRANSFER_TO = __ENV.TRANSFER_TO || '6543210987654321';
const MERCHANT_ID = Number(__ENV.MERCHANT_ID || 1);
const TOPUP_METHOD = __ENV.TOPUP_METHOD || 'bca';
const AMOUNT = Number(__ENV.AMOUNT || 50000);

export const options = {
  scenarios: {
    financial_mutation: {
      // A handful of VUs: every iteration performs the full mutation sequence.
      executor: 'constant-vus',
      vus: Number(__ENV.VUS || 2),
      duration: __ENV.DURATION || '1m',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
  },
};

function headers(key, extra = {}) {
  return {
    headers: Object.assign(
      {
        Authorization: TOKEN,
        'Content-Type': 'application/json',
        'Idempotency-Key': key,
      },
      extra,
    ),
  };
}

export default function () {
  // Unique suffix per VU+iteration so each pass exercises fresh idempotency keys.
  const suffix = `${__VU}-${__ITER}-${Date.now()}`;
  const key = `mutation-${suffix}`;

  // ---- 1. Topup create with a key ----
  const topupBody = JSON.stringify({
    card_number: CARD_NUMBER,
    topup_amount: AMOUNT,
    topup_method: TOPUP_METHOD,
  });
  const first = http.post(`${BASE_URL}/api/topup-command/create`, topupBody, headers(key));
  check(first, {
    'topup create is 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  const firstID = first.json('data.id');

  // ---- 2. Retry same key + payload -> replay, not a second topup ----
  if (firstID !== undefined && firstID !== null) {
    const replay = http.post(`${BASE_URL}/api/topup-command/create`, topupBody, headers(key));
    check(replay, {
      'idempotent retry replays with same id': (r) => r.json('data.id') === firstID,
      'idempotent retry is 2xx': (r) => r.status >= 200 && r.status < 300,
    });
  } else {
    check(null, { 'topup returned an id (data.id)': false });
  }

  // ---- 3. Same key + different payload -> 409 conflict ----
  const conflictBody = JSON.stringify({
    card_number: CARD_NUMBER,
    topup_amount: AMOUNT + 1,
    topup_method: TOPUP_METHOD,
  });
  const conflict = http.post(`${BASE_URL}/api/topup-command/create`, conflictBody, headers(key));
  check(conflict, {
    'same key different payload -> 409': (r) => r.status === 409,
  });

  // ---- 4. Invalid amounts -> 400 (P1.7) ----
  for (const badAmount of [0, -1]) {
    const bad = http.post(
      `${BASE_URL}/api/topup-command/create`,
      JSON.stringify({ card_number: CARD_NUMBER, topup_amount: badAmount, topup_method: TOPUP_METHOD }),
      headers(`mutation-bad-${badAmount}-${suffix}`),
    );
    check(bad, {
      [`topup amount ${badAmount} -> 400`]: (r) => r.status === 400,
    });
  }

  // ---- 5. Withdraw amount 0 -> 400 (P1.7) ----
  const badWithdraw = http.post(
    `${BASE_URL}/api/withdraw-command/create`,
    JSON.stringify({ card_number: CARD_NUMBER, withdraw_amount: 0, withdraw_time: new Date().toISOString() }),
    headers(`mutation-withdraw0-${suffix}`),
  );
  check(badWithdraw, {
    'withdraw amount 0 -> 400': (r) => r.status === 400,
  });

  // ---- 6. Transfer to self -> 400 (P1.7) ----
  const selfTransfer = http.post(
    `${BASE_URL}/api/transfer-command/create`,
    JSON.stringify({ transfer_from: CARD_NUMBER, transfer_to: CARD_NUMBER, transfer_amount: AMOUNT }),
    headers(`mutation-self-${suffix}`),
  );
  check(selfTransfer, {
    'transfer to self -> 400': (r) => r.status === 400,
  });

  // ---- 7. Transaction with absurd amount -> 409 insufficient (P1.5) ----
  const hugeTx = http.post(
    `${BASE_URL}/api/transaction-command/create`,
    JSON.stringify({
      card_number: CARD_NUMBER,
      amount: 999999999999,
      payment_method: 'visa',
      merchant_id: MERCHANT_ID,
      transaction_time: new Date().toISOString(),
    }),
    headers(`mutation-huge-${suffix}`, { 'X-API-Key': API_KEY }),
  );
  check(hugeTx, {
    'huge transaction -> 409 conflict': (r) => r.status === 409,
    'huge transaction never 500': (r) => r.status !== 500,
  });

  sleep(0.2);
}
