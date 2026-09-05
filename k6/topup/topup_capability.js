import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const TOPUP_ID = 1;
const CARD_NUMBER = '1234567890123456';

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
    `/api/topups?page=1&page_size=10`,
    `/api/topups/active?page=1&page_size=10`,
    `/api/topups/trashed?page=1&page_size=10`,
    `/api/topups/${TOPUP_ID}`,
    `/api/topups/card-number/${CARD_NUMBER}`,
    `/api/stats/topup/status/success/monthly?year=2024&month=1`,
    `/api/stats/topup/status/success/yearly?year=2024`,
    `/api/stats/topup/status/failed/monthly?year=2024&month=1`,
    `/api/stats/topup/status/failed/yearly?year=2024`,
    `/api/stats/topup/method/monthly?year=2024`,
    `/api/stats/topup/method/yearly?year=2024`,
    `/api/stats/topup/amount/monthly?year=2024`,
    `/api/stats/topup/amount/yearly?year=2024`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
