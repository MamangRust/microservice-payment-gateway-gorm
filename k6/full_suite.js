import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:5000';
const VUS = Number(__ENV.K6_VUS || 20);
const DURATION = __ENV.K6_DURATION || '8m';

export const options = {
  scenarios: {
    all_endpoints_smoke: {
      executor: 'constant-vus',
      vus: VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
  },
};

function requestParams(token) {
  return { headers: { Authorization: `Bearer ${token}` } };
}

function responseItems(response) {
  if (response.status !== 200) {
    return [];
  }

  const items = response.json('data');
  return Array.isArray(items) ? items : [];
}

function fetchItems(path, token) {
  return responseItems(http.get(`${BASE_URL}${path}`, requestParams(token)));
}

// Login once in setup() and discover fixtures from the running API. This avoids
// repeatedly requesting stale IDs/card numbers that are not present in local data.
export function setup() {
  const loginResponse = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify({ email: 'john.doe@hellodota.com', password: 'password123' }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  const token = loginResponse.json('data.access_token');
  if (!token) {
    throw new Error(`login failed: ${loginResponse.status} ${loginResponse.body}`);
  }

  const userItems = fetchItems('/api/user-query?page=1&page_size=10', token);
  const roleItems = fetchItems('/api/role-query?page=1&page_size=10', token);
  const merchantItems = fetchItems('/api/merchant-query?page=1&page_size=10', token);
  const saldoItems = fetchItems('/api/saldo-query?page=1&page_size=10', token);
  const topupItems = fetchItems('/api/topup-query?page=1&page_size=10', token);
  const transactionItems = fetchItems('/api/transaction-query?page=1&page_size=10', token);
  const transferItems = fetchItems('/api/transfer-query?page=1&page_size=10', token);
  const withdrawItems = fetchItems('/api/withdraw-query?page=1&page_size=10', token);
  const cardItems = fetchItems('/api/card-query?page=1&page_size=10', token);

  return {
    token,
    userId: userItems[0]?.id,
    roleId: roleItems[0]?.id,
    merchantId: merchantItems[0]?.id,
    saldo: saldoItems[0],
    topup: topupItems[0],
    transaction: transactionItems[0],
    transfer: transferItems[0],
    withdraw: withdrawItems[0],
    card: cardItems[0],
  };
}

export default function (data) {
  const params = requestParams(data.token);
  const endpoints = [
    // ===== auth =====
    '/api/auth/hello',
    '/api/auth/me',

    // ===== user =====
    '/api/user-query?page=1&page_size=10',
    '/api/user-query/active?page=1&page_size=10',
    '/api/user-query/trashed?page=1&page_size=10',

    // ===== role =====
    '/api/role-query?page=1&page_size=10',
    '/api/role-query/active?page=1&page_size=10',
    '/api/role-query/trashed?page=1&page_size=10',

    // ===== merchant =====
    '/api/merchant-query?page=1&page_size=10',
    '/api/merchant-query/active?page=1&page_size=10',
    '/api/merchant-query/trashed?page=1&page_size=10',
    '/api/merchant/stats/transactions?page=1&page_size=10',
    '/api/merchant/stats/amount/monthly?year=2024',

    // ===== saldo =====
    '/api/saldo-query?page=1&page_size=10',
    '/api/saldo-query/active?page=1&page_size=10',
    '/api/saldo-query/trashed?page=1&page_size=10',
    '/api/saldo/stats/total/monthly?year=2024&month=1',
    '/api/saldo/stats/total/yearly?year=2024',
    '/api/saldo/stats/balance/monthly?year=2024',
    '/api/saldo/stats/balance/yearly?year=2024',

    // ===== topup =====
    '/api/topup-query?page=1&page_size=10',
    '/api/topup-query/active?page=1&page_size=10',
    '/api/topup-query/trashed?page=1&page_size=10',
    '/api/topup/stats/status/monthly/success?year=2024&month=1',
    '/api/topup/stats/status/yearly/success?year=2024',
    '/api/topup/stats/status/monthly/failed?year=2024&month=1',
    '/api/topup/stats/status/yearly/failed?year=2024',
    '/api/topup/stats/method/monthly?year=2024',
    '/api/topup/stats/method/yearly?year=2024',
    '/api/topup/stats/amount/monthly?year=2024',
    '/api/topup/stats/amount/yearly?year=2024',

    // ===== transaction =====
    '/api/transaction-query?page=1&page_size=10',
    '/api/transaction-query/active?page=1&page_size=10',
    '/api/transaction-query/trashed?page=1&page_size=10',
    '/api/transaction/stats/status/monthly/success?year=2024&month=1',
    '/api/transaction/stats/status/yearly/success?year=2024',
    '/api/transaction/stats/status/monthly/failed?year=2024&month=1',
    '/api/transaction/stats/status/yearly/failed?year=2024',
    '/api/transaction/stats/method/monthly?year=2024',
    '/api/transaction/stats/method/yearly?year=2024',
    '/api/transaction/stats/amount/monthly?year=2024',
    '/api/transaction/stats/amount/yearly?year=2024',

    // ===== transfer =====
    '/api/transfer-query?page=1&page_size=10',
    '/api/transfer-query/active?page=1&page_size=10',
    '/api/transfer-query/trashed?page=1&page_size=10',
    '/api/transfer/stats/status/monthly/success?year=2024&month=1',
    '/api/transfer/stats/status/yearly/success?year=2024',
    '/api/transfer/stats/status/monthly/failed?year=2024&month=1',
    '/api/transfer/stats/status/yearly/failed?year=2024',
    '/api/transfer/stats/amount/monthly?year=2024',
    '/api/transfer/stats/amount/yearly?year=2024',

    // ===== withdraw =====
    '/api/withdraw-query?page=1&page_size=10',
    '/api/withdraw-query/active?page=1&page_size=10',
    '/api/withdraw-query/trashed?page=1&page_size=10',
    '/api/withdraw/stats/status/monthly/success?year=2024&month=1',
    '/api/withdraw/stats/status/yearly/success?year=2024',
    '/api/withdraw/stats/status/monthly/failed?year=2024&month=1',
    '/api/withdraw/stats/status/yearly/failed?year=2024',
    '/api/withdraw/stats/amount/monthly?year=2024',
    '/api/withdraw/stats/amount/yearly?year=2024',

    // ===== card query and stats =====
    '/api/card-query?page=1&page_size=10',
    '/api/card-query/active?page=1&page_size=10',
    '/api/card-query/user',
    '/api/card/stats/balance/monthly?year=2025',
    '/api/card/stats/balance/yearly?year=2025',
    '/api/card/stats/topup/monthly?year=2025',
    '/api/card/stats/topup/yearly?year=2025',
    '/api/card/stats/transaction/monthly?year=2025',
    '/api/card/stats/transaction/yearly?year=2025',
    '/api/card/stats/transfer/sender/monthly?year=2025',
    '/api/card/stats/transfer/receiver/monthly?year=2025',
    '/api/card/stats/transfer/sender/yearly?year=2025',
    '/api/card/stats/transfer/receiver/yearly?year=2025',
    '/api/card/stats/withdraw/monthly?year=2025',
    '/api/card/stats/withdraw/yearly?year=2025',
    '/api/card-query/trashed?page=1&page_size=10',
  ];

  // Add detail endpoints only when setup discovered a real fixture.
  if (data.userId !== undefined) {
    endpoints.push(`/api/user-query/${data.userId}`);
    endpoints.push(`/api/role-query/user/${data.userId}`);
  }
  if (data.roleId !== undefined) {
    endpoints.push(`/api/role-query/${data.roleId}`);
  }
  if (data.merchantId !== undefined) {
    endpoints.push(`/api/merchant-query/${data.merchantId}`);
  }

  const cardNumber = data.card?.card_number;
  if (data.card?.id !== undefined) {
    endpoints.push(`/api/card-query/${data.card.id}`);
  }
  if (data.saldo?.id !== undefined) {
    endpoints.push(`/api/saldo-query/${data.saldo.id}`);
  }
  if (data.saldo?.card_number) {
    endpoints.push(`/api/saldo-query/card_number/${data.saldo.card_number}`);
  }
  if (data.topup?.id !== undefined) {
    endpoints.push(`/api/topup-query/${data.topup.id}`);
  }
  if (data.topup?.card_number) {
    endpoints.push(`/api/topup-query/card-number/${data.topup.card_number}`);
  }
  if (data.transaction?.id !== undefined) {
    endpoints.push(`/api/transaction-query/${data.transaction.id}`);
  }
  if (data.transaction?.card_number) {
    endpoints.push(`/api/transaction-query/card-number/${data.transaction.card_number}`);
  }
  if (data.transfer?.id !== undefined) {
    endpoints.push(`/api/transfer-query/${data.transfer.id}`);
  }
  if (data.transfer?.transfer_from) {
    endpoints.push(`/api/transfer-query/transfer_from/${data.transfer.transfer_from}`);
  }
  if (data.transfer?.transfer_to) {
    endpoints.push(`/api/transfer-query/transfer_to/${data.transfer.transfer_to}`);
  }
  if (data.withdraw?.id !== undefined) {
    endpoints.push(`/api/withdraw-query/${data.withdraw.id}`);
  }
  if (data.withdraw?.card_number) {
    endpoints.push(`/api/withdraw-query/card-number/${data.withdraw.card_number}`);
  }
  if (cardNumber) {
    endpoints.push(`/api/card/stats/balance/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/balance/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/topup/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/topup/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transaction/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transaction/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transfer/sender/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transfer/receiver/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transfer/sender/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/transfer/receiver/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/withdraw/monthly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card/stats/withdraw/yearly/${cardNumber}?year=2025`);
    endpoints.push(`/api/card-query/card_number/${cardNumber}`);
  }

  for (const endpoint of endpoints) {
    const response = http.get(`${BASE_URL}${endpoint}`, params);
    check(response, { [`GET ${endpoint} success`]: (r) => r.status === 200 });
  }

  sleep(0.05);
}
