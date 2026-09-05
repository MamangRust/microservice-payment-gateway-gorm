import http from 'k6/http';
import { sleep } from 'k6';

const BASE_URL = 'http://localhost:5000';

// 1 VU, 1 iteration — sequential per-endpoint timing (cold cache)
export const options = {
  vus: 1,
  iterations: 1,
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

const CARD = '4415599957419074';
const CARD_OLD = '4016531504435114';

const endpoints = [
  // ===== merchant =====
  '/api/merchants/monthly-amount?year=2024',
  // ===== saldo =====
  '/api/saldos/monthly-total-balance?year=2024&month=1',
  '/api/saldos/yearly-total-balance?year=2024',
  '/api/saldos/monthly-balances?year=2024',
  '/api/saldos/yearly-balances?year=2024',
  // ===== topup =====
  '/api/topups/monthly-success?year=2024&month=1',
  '/api/topups/yearly-success?year=2024',
  '/api/topups/monthly-failed?year=2024&month=1',
  '/api/topups/yearly-failed?year=2024',
  '/api/topups/monthly-methods?year=2024',
  '/api/topups/yearly-methods?year=2024',
  '/api/topups/monthly-amounts?year=2024',
  '/api/topups/yearly-amounts?year=2024',
  // ===== transaction =====
  '/api/transactions/monthly-success?year=2024&month=1',
  '/api/transactions/yearly-success?year=2024',
  '/api/transactions/monthly-failed?year=2024&month=1',
  '/api/transactions/yearly-failed?year=2024',
  '/api/transactions/monthly-methods?year=2024',
  '/api/transactions/yearly-methods?year=2024',
  '/api/transactions/monthly-amounts?year=2024',
  '/api/transactions/yearly-amounts?year=2024',
  // ===== transfer =====
  '/api/transfers/monthly-success?year=2024&month=1',
  '/api/transfers/yearly-success?year=2024',
  '/api/transfers/monthly-failed?year=2024&month=1',
  '/api/transfers/yearly-failed?year=2024',
  '/api/transfers/monthly-amounts?year=2024',
  '/api/transfers/yearly-amounts?year=2024',
  // ===== withdraw =====
  '/api/withdraws/monthly-success?year=2024&month=1',
  '/api/withdraws/yearly-success?year=2024',
  '/api/withdraws/monthly-failed?year=2024&month=1',
  '/api/withdraws/yearly-failed?year=2024',
  '/api/withdraws/monthly-amount?year=2024',
  '/api/withdraws/yearly-amount?year=2024',
  // ===== card stats (new) =====
  '/api/cards/stats/balance/monthly?year=2025',
  '/api/cards/stats/balance/yearly?year=2025',
  `/api/cards/stats/balance/monthly/by-card?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/balance/yearly/by-card?year=2025&card_number=${CARD}`,
  '/api/cards/stats/topup/monthly?year=2025',
  '/api/cards/stats/topup/yearly?year=2025',
  `/api/cards/stats/topup/monthly/by-card?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/topup/yearly/by-card?year=2025&card_number=${CARD}`,
  '/api/cards/stats/transaction/monthly?year=2025',
  '/api/cards/stats/transaction/yearly?year=2025',
  `/api/cards/stats/transaction/monthly/by-card?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/transaction/yearly/by-card?year=2025&card_number=${CARD}`,
  '/api/cards/stats/transfer/monthly/sender?year=2025',
  '/api/cards/stats/transfer/monthly/receiver?year=2025',
  '/api/cards/stats/transfer/yearly/sender?year=2025',
  '/api/cards/stats/transfer/yearly/receiver?year=2025',
  `/api/cards/stats/transfer/monthly/by-card/sender?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/transfer/monthly/by-card/receiver?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/transfer/yearly/by-card/sender?year=2025&card_number=${CARD}`,
  `/api/cards/stats/transfer/yearly/by-card/receiver?year=2025&card_number=${CARD}`,
  '/api/cards/stats/withdraw/monthly?year=2025',
  '/api/cards/stats/withdraw/yearly?year=2025',
  `/api/cards/stats/withdraw/monthly/by-card?year=2025&month=1&card_number=${CARD}`,
  `/api/cards/stats/withdraw/yearly/by-card?year=2025&card_number=${CARD}`,
  // ===== legacy stats-ish card paths =====
  `/api/card/card_number/${CARD_OLD}`,
];

export default function (data) {
  const params = { headers: { Authorization: `Bearer ${data.token}` } };
  const results = [];

  for (let endpoint of endpoints) {
    const res = http.get(`${BASE_URL}${endpoint}`, params);
    const dur = res.timings.duration;
    results.push({ endpoint, dur, status: res.status });
    console.log(`${String(dur.toFixed(1)).padStart(10)} ms  HTTP ${res.status}  ${endpoint}`);
    if (dur > 1000) console.log(`  >>> SLOW (${(dur / 1000).toFixed(2)}s)`);
    sleep(0.2);
  }

  results.sort((a, b) => b.dur - a.dur);
  console.log('\n===== TOP 15 SLOWEST (cold cache) =====');
  for (let i = 0; i < Math.min(15, results.length); i++) {
    const r = results[i];
    console.log(`${String((r.dur / 1000).toFixed(2)).padStart(8)}s  HTTP ${r.status}  ${r.endpoint}`);
  }
  console.log(`\nTotal endpoints: ${results.length}`);
}
