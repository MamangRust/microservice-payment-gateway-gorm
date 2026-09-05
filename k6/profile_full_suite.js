import http from 'k6/http';
import { sleep } from 'k6';
import { Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:80';

// Replicates the EXACT full_suite conditions (constant-vus 20, 8m) so the
// 5-min Redis TTL expiry stampede that produced the 27.98s outlier can
// reproduce — but with a per-URL Trend metric per endpoint so we can rank
// max latency per endpoint afterwards (instead of relying on summary-export).
export const options = {
  scenarios: {
    all_endpoints_smoke: {
      executor: 'constant-vus',
      vus: 20,
      duration: '5m',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.1'],
  },
};

export function setup() {
  const res = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify({ email: 'john.doe@hellodota.com', password: 'password123' }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  const token = res.json('data.access_token');
  if (!token) throw new Error(`login failed: ${res.status} ${res.body}`);
  return { token };
}

const USER_ID = 1;
const ROLE_ID = 1;
const MERCHANT_ID = 1;
const SALDO_ID = 1;
const TOPUP_ID = 1;
const TRANSACTION_ID = 1;
const TRANSFER_ID = 1;
const WITHDRAW_ID = 1;
const CARD_NUMBER = '4415599957419074';
const CARD_NUMBER_OLD = '4016531504435114';

const endpoints = [
  // ===== auth =====
  '/api/auth/hello',
  '/api/auth/me',
  // ===== user =====
  '/api/user?page=1&limit=10',
  `/api/user/${USER_ID}`,
  '/api/user/active?page=1&limit=10',
  '/api/user/trashed?page=1&limit=10',
  // ===== role =====
  '/api/role?page=1&limit=10',
  '/api/role/active?page=1&limit=10',
  '/api/role/trashed?page=1&limit=10',
  `/api/role/${ROLE_ID}`,
  `/api/role/user/${USER_ID}`,
  // ===== merchant =====
  '/api/merchants?page=1&page_size=10',
  '/api/merchants/active?page=1&page_size=10',
  '/api/merchants/trashed?page=1&page_size=10',
  `/api/merchants/${MERCHANT_ID}`,
  '/api/merchants/transactions?page=1&page_size=10',
  '/api/merchants/monthly-amount?year=2024',
  // ===== saldo =====
  '/api/saldos?page=1&page_size=10',
  '/api/saldos/active?page=1&page_size=10',
  '/api/saldos/trashed?page=1&page_size=10',
  `/api/saldos/${SALDO_ID}`,
  `/api/saldos/card-number/${CARD_NUMBER}`,
  `/api/saldos/user/${USER_ID}`,
  '/api/saldos/monthly-total-balance?year=2024&month=1',
  '/api/saldos/yearly-total-balance?year=2024',
  '/api/saldos/monthly-balances?year=2024',
  '/api/saldos/yearly-balances?year=2024',
  // ===== topup =====
  '/api/topups?page=1&page_size=10',
  '/api/topups/active?page=1&page_size=10',
  '/api/topups/trashed?page=1&page_size=10',
  `/api/topups/${TOPUP_ID}`,
  `/api/topups/card-number/${CARD_NUMBER}`,
  '/api/topups/monthly-success?year=2024&month=1',
  '/api/topups/yearly-success?year=2024',
  '/api/topups/monthly-failed?year=2024&month=1',
  '/api/topups/yearly-failed?year=2024',
  '/api/topups/monthly-methods?year=2024',
  '/api/topups/yearly-methods?year=2024',
  '/api/topups/monthly-amounts?year=2024',
  '/api/topups/yearly-amounts?year=2024',
  // ===== transaction =====
  '/api/transactions?page=1&page_size=10',
  '/api/transactions/active?page=1&page_size=10',
  '/api/transactions/trashed?page=1&page_size=10',
  `/api/transactions/${TRANSACTION_ID}`,
  `/api/transactions/card-number/${CARD_NUMBER}`,
  '/api/transactions/monthly-success?year=2024&month=1',
  '/api/transactions/yearly-success?year=2024',
  '/api/transactions/monthly-failed?year=2024&month=1',
  '/api/transactions/yearly-failed?year=2024',
  '/api/transactions/monthly-methods?year=2024',
  '/api/transactions/yearly-methods?year=2024',
  '/api/transactions/monthly-amounts?year=2024',
  '/api/transactions/yearly-amounts?year=2024',
  // ===== transfer =====
  '/api/transfers?page=1&page_size=10',
  '/api/transfers/active?page=1&page_size=10',
  '/api/transfers/trashed?page=1&page_size=10',
  `/api/transfers/${TRANSFER_ID}`,
  '/api/transfers/monthly-success?year=2024&month=1',
  '/api/transfers/yearly-success?year=2024',
  '/api/transfers/monthly-failed?year=2024&month=1',
  '/api/transfers/yearly-failed?year=2024',
  '/api/transfers/monthly-amounts?year=2024',
  '/api/transfers/yearly-amounts?year=2024',
  // ===== withdraw =====
  '/api/withdraws?page=1&page_size=10',
  '/api/withdraws/active?page=1&page_size=10',
  '/api/withdraws/trashed?page=1&page_size=10',
  `/api/withdraws/${WITHDRAW_ID}`,
  `/api/withdraws/card-number/${CARD_NUMBER}`,
  '/api/withdraws/monthly-success?year=2024&month=1',
  '/api/withdraws/yearly-success?year=2024',
  '/api/withdraws/monthly-failed?year=2024&month=1',
  '/api/withdraws/yearly-failed?year=2024',
  '/api/withdraws/monthly-amount?year=2024',
  '/api/withdraws/yearly-amount?year=2024',
  // ===== card (new paths) =====
  '/api/cards',
  '/api/cards/active?page=1&limit=10',
  '/api/cards/by-user/1',
  '/api/cards/stats/balance/monthly?year=2025',
  '/api/cards/stats/balance/yearly?year=2025',
  `/api/cards/stats/balance/monthly/by-card?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/balance/yearly/by-card?year=2025&card_number=${CARD_NUMBER}`,
  '/api/cards/stats/topup/monthly?year=2025',
  '/api/cards/stats/topup/yearly?year=2025',
  `/api/cards/stats/topup/monthly/by-card?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/topup/yearly/by-card?year=2025&card_number=${CARD_NUMBER}`,
  '/api/cards/stats/transaction/monthly?year=2025',
  '/api/cards/stats/transaction/yearly?year=2025',
  `/api/cards/stats/transaction/monthly/by-card?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/transaction/yearly/by-card?year=2025&card_number=${CARD_NUMBER}`,
  '/api/cards/stats/transfer/monthly/sender?year=2025',
  '/api/cards/stats/transfer/monthly/receiver?year=2025',
  '/api/cards/stats/transfer/yearly/sender?year=2025',
  '/api/cards/stats/transfer/yearly/receiver?year=2025',
  `/api/cards/stats/transfer/monthly/by-card/sender?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/transfer/monthly/by-card/receiver?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/transfer/yearly/by-card/sender?year=2025&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/transfer/yearly/by-card/receiver?year=2025&card_number=${CARD_NUMBER}`,
  '/api/cards/stats/withdraw/monthly?year=2025',
  '/api/cards/stats/withdraw/yearly?year=2025',
  `/api/cards/stats/withdraw/monthly/by-card?year=2025&month=1&card_number=${CARD_NUMBER}`,
  `/api/cards/stats/withdraw/yearly/by-card?year=2025&card_number=${CARD_NUMBER}`,
  // ===== card (legacy paths) =====
  '/api/card?page=1&limit=10',
  '/api/card/active?page=1&limit=10',
  '/api/card/trashed?page=1&limit=10',
  `/api/card/user?user_id=${USER_ID}`,
  `/api/card/card_number/${CARD_NUMBER_OLD}`,
  `/api/card/${USER_ID}`,
];

// One Trend metric per endpoint so handleSummary can rank max latency per URL.
const trends = endpoints.map((_, i) => new Trend(`url_${i}`, true));

export default function (data) {
  const params = { headers: { Authorization: `Bearer ${data.token}` } };
  for (let i = 0; i < endpoints.length; i++) {
    const res = http.get(`${BASE_URL}${endpoints[i]}`, params);
    trends[i].add(res.timings.duration);
  }
  sleep(0.05);
}

export function handleSummary(data) {
  const rows = [];
  for (let i = 0; i < trends.length; i++) {
    const m = data.metrics[`url_${i}`];
    if (!m || !m.values) continue;
    rows.push({ ep: endpoints[i], max: m.values.max, avg: m.values.avg, count: m.values.count });
  }
  rows.sort((a, b) => b.max - a.max);

  let out = '\n===== SLOWEST PER-URL (8m, 20 VUs, full endpoint set) =====\n';
  for (const r of rows.slice(0, 20)) {
    out += `${String(r.max.toFixed(1)).padStart(10)}ms max | avg ${r.avg.toFixed(1)}ms | n=${r.count} | ${r.ep}\n`;
  }
  const dur = data.metrics.http_req_duration && data.metrics.http_req_duration.values;
  if (dur) {
    out += `\noverall http_req_duration: avg=${dur.avg.toFixed(2)}ms max=${dur.max.toFixed(2)}ms p95=${(dur['p(95)'] || 0).toFixed(2)}ms\n`;
    out += `total reqs: ${(data.metrics.http_reqs && data.metrics.http_reqs.values.count) || '?'}\n`;
  }
  return { stdout: out };
}
