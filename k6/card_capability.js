import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN =
  'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk ';

export const options = {
  scenarios: {
    scalability_test: {
      executor: 'ramping-arrival-rate',
      startRate: 50,
      timeUnit: '1s',
      stages: [
        { duration: '1m', target: 100 },
        { duration: '1m', target: 300 },
        { duration: '1m', target: 600 },
      ],
      preAllocatedVUs: 100,
      maxVUs: 900,
    },
  },
};

export default function () {
  const params = { headers: { Authorization: TOKEN } };

  const basicEndpoints = [
    // ===== card =====
    '/api/card?page=1&limit=10',
    '/api/card/active?page=1&limit=10',
    '/api/card/trashed?page=1&limit=10',
    '/api/card/user?user_id=11',
    '/api/card/card_number/4389195775213720',
    '/api/card/11',

    // ===== Dashboard =====
    '/api/stats/card/dashboard',
    '/api/stats/card/dashboard/4016531504435114',

    // ===== Balance =====
    '/api/stats/card/balance/monthly?year=2025&month=1',
    '/api/stats/card/balance/yearly?year=2025',
    '/api/stats/card/balance/monthly/4016531504435114?year=2025',
    '/api/stats/card/balance/yearly/4016531504435114?year=2025',

    // ===== Topup =====
    '/api/stats/card/topup/monthly?year=2025&month=1',
    '/api/stats/card/topup/yearly?year=2025',
    '/api/stats/card/topup/monthly/4016531504435114?year=2025',
    '/api/stats/card/topup/yearly/4016531504435114?year=2025',

    // ===== Withdraw =====
    '/api/stats/card/withdraw/monthly?year=2025&month=1',
    '/api/stats/card/withdraw/yearly?year=2025',
    '/api/stats/card/withdraw/monthly/4016531504435114?year=2025',
    '/api/stats/card/withdraw/yearly/4016531504435114?year=2025',

    // ===== Transaction =====
    '/api/stats/card/transaction/monthly?year=2025&month=1',
    '/api/stats/card/transaction/yearly?year=2025',
    '/api/stats/card/transaction/monthly/4016531504435114?year=2025',
    '/api/stats/card/transaction/yearly/4016531504435114?year=2025',

    // ===== Transfer Sender =====
    '/api/stats/card/transfer/sender/monthly?year=2025&month=1',
    '/api/stats/card/transfer/sender/yearly?year=2025',
    '/api/stats/card/transfer/sender/monthly/4016531504435114?year=2025',
    '/api/stats/card/transfer/sender/yearly/4016531504435114?year=2025',

    // ===== Transfer Receiver =====
    '/api/stats/card/transfer/receiver/monthly?year=2025&month=1',
    '/api/stats/card/transfer/receiver/yearly?year=2025',
    '/api/stats/card/transfer/receiver/monthly/4016531504435114?year=2025',
    '/api/stats/card/transfer/receiver/yearly/4016531504435114?year=2025',
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
