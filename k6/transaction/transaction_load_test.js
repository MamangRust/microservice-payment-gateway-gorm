import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const TRANSACTION_ID = 1;
const CARD_NUMBER = '1234567890123456';

export const options = {
  scenarios: {
    load_test: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '2m',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<200'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const params = { headers: { Authorization: TOKEN } };

  const basicEndpoints = [
    `/api/transactions?page=1&page_size=10`,
    `/api/transactions/active?page=1&page_size=10`,
    `/api/transactions/trashed?page=1&page_size=10`,
    `/api/transactions/${TRANSACTION_ID}`,
    `/api/transactions/card-number/${CARD_NUMBER}`,
    `/api/stats/transaction/status/success/monthly?year=2024&month=1`,
    `/api/stats/transaction/status/success/yearly?year=2024`,
    `/api/stats/transaction/status/failed/monthly?year=2024&month=1`,
    `/api/stats/transaction/status/failed/yearly?year=2024`,
    `/api/stats/transaction/method/monthly?year=2024`,
    `/api/stats/transaction/method/yearly?year=2024`,
    `/api/stats/transaction/amount/monthly?year=2024`,
    `/api/stats/transaction/amount/yearly?year=2024`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
