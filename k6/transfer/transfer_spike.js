import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const TRANSFER_ID = 1;

export const options = {
  scenarios: {
    spike_test: {
      executor: 'ramping-vus',
      stages: [
        { duration: '10s', target: 50 },
        { duration: '10s', target: 1000 },
        { duration: '30s', target: 1000 },
        { duration: '10s', target: 50 },
      ],
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
    `/api/transfers/monthly-success?year=2024&month=1`,
    `/api/transfers/yearly-success?year=2024`,
    `/api/transfers/monthly-failed?year=2024&month=1`,
    `/api/transfers/yearly-failed?year=2024`,
    `/api/transfers/monthly-amounts?year=2024`,
    `/api/transfers/yearly-amounts?year=2024`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
