import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const TRANSFER_ID = 1;

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
    `/api/transfers?page=1&page_size=10`,
    `/api/transfers/active?page=1&page_size=10`,
    `/api/transfers/trashed?page=1&page_size=10`,
    `/api/transfers/${TRANSFER_ID}`,
    `/api/stats/transfer/status/success/monthly?year=2024&month=1`,
    `/api/stats/transfer/status/success/yearly?year=2024`,
    `/api/stats/transfer/status/failed/monthly?year=2024&month=1`,
    `/api/stats/transfer/status/failed/yearly?year=2024`,
    `/api/stats/transfer/amount/monthly?year=2024`,
    `/api/stats/transfer/amount/yearly?year=2024`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
