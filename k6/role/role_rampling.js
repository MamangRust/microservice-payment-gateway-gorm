import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN =
  'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const USER_ID = 1;
const ROLE_ID = 1;

export const options = {
  scenarios: {
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 100,
      stages: [
        { duration: '30s', target: 300 },
        { duration: '30s', target: 600 },
        { duration: '30s', target: 1000 },
        { duration: '30s', target: 1500 },
      ],
    },
  },
};

export default function () {
  const params = { headers: { Authorization: TOKEN } };

  const basicEndpoints = [
    `/api/role?page=1&limit=10`,
    `/api/role/active?page=1&limit=10`,
    `/api/role/trashed?page=1&limit=10`,
    `/api/role/${ROLE_ID}`,
    `/api/role/user/${USER_ID}`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
