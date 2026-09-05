// Financial command concurrency/load test (Phase 5).
//
// Exercises the four money-moving creates through the API Gateway with the
// required Idempotency-Key header. Every iteration uses a unique key, so a
// retry with the same key must replay instead of double-moving money.
//
// Configuration via environment variables (k6 -e KEY=VALUE):
//   BASE_URL          default http://localhost:5000
//   TOKEN             Bearer token (required)
//   API_KEY           merchant API key for transaction (optional)
//   CARD_NUMBER       default 1234567890123456
//   TRANSFER_TO       receiver card (default 6543210987654321)
//   MERCHANT_ID       merchant id for transaction (default 1)
//   TOPUP_METHOD      valid method (default bca; see pkg/method_topup)
//   VUS               default 100
//   DURATION          default 2m
//   AMOUNT            default 50000 (minimum enforced by topup/withdraw validation)
//
// Run: k6 run -e TOKEN='Bearer ...' k6/financial/financial_concurrency.js

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
    financial_commands: {
      executor: 'constant-vus',
      vus: Number(__ENV.VUS || 100),
      duration: __ENV.DURATION || '2m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

function uniqueKey(prefix) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

function commandParams(key) {
  return {
    headers: {
      Authorization: TOKEN,
      'Content-Type': 'application/json',
      'Idempotency-Key': key,
    },
  };
}

export default function () {
  const action = (__VU + __ITER) % 4;

  if (action === 0) {
    const body = JSON.stringify({
      card_number: CARD_NUMBER,
      topup_amount: AMOUNT,
      topup_method: TOPUP_METHOD,
    });
    const res = http.post(`${BASE_URL}/api/topup-command/create`, body, commandParams(uniqueKey('topup')));
    check(res, {
      'topup create accepted': (r) => r.status === 200 || r.status === 409,
      'topup no 5xx': (r) => r.status < 500,
    });
  } else if (action === 1) {
    const body = JSON.stringify({
      card_number: CARD_NUMBER,
      amount: AMOUNT,
      payment_method: 'visa',
      merchant_id: MERCHANT_ID,
      transaction_time: new Date().toISOString(),
    });
    const res = http.post(
      `${BASE_URL}/api/transaction-command/create`,
      body,
      {
        headers: Object.assign(commandParams(uniqueKey('tx')).headers, { 'X-API-Key': API_KEY }),
      },
    );
    check(res, {
      'transaction create accepted': (r) => r.status === 200 || r.status === 409,
      'transaction no 5xx': (r) => r.status < 500,
    });
  } else if (action === 2) {
    const body = JSON.stringify({
      transfer_from: CARD_NUMBER,
      transfer_to: TRANSFER_TO,
      transfer_amount: AMOUNT,
    });
    const res = http.post(`${BASE_URL}/api/transfer-command/create`, body, commandParams(uniqueKey('transfer')));
    check(res, {
      'transfer create accepted': (r) => r.status === 200 || r.status === 409,
      'transfer no 5xx': (r) => r.status < 500,
    });
  } else {
    const body = JSON.stringify({
      card_number: CARD_NUMBER,
      withdraw_amount: AMOUNT,
      withdraw_time: new Date().toISOString(),
    });
    const res = http.post(`${BASE_URL}/api/withdraw-command/create`, body, commandParams(uniqueKey('withdraw')));
    check(res, {
      'withdraw create accepted': (r) => r.status === 200 || r.status === 409,
      'withdraw no 5xx': (r) => r.status < 500,
    });
  }

  sleep(0.1);
}
