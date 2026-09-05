import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';
const TOKEN =
  'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiYXVkIjpbImFjY2VzcyJdLCJleHAiOjE3ODU1NTY2MTJ9.O2__bTSOSd-SA0yX5Girec-v4VGkrj3rgwmrImocCdk';

const USER_ID = 41;

export default function () {
  const params = {
    headers: { Authorization: TOKEN, 'Content-Type': 'application/json' },
  };

  const basicEndpoints = [
    `/api/user?page=1&limit=10`,
    `/api/user/${USER_ID}`,
    `/api/user/active?page=1&limit=10`,
    `/api/user/trashed?page=1&limit=10`,
  ];

  for (let endpoint of basicEndpoints) {
    let res = http.get(`${BASE_URL}${endpoint}`, params);
    check(res, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.1);
}
