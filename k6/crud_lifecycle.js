import http from 'k6/http';
import { check, fail } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:5000';
const VUS = Number(__ENV.K6_VUS || 1);
const DURATION = __ENV.K6_DURATION || '30s';

export const options = {
  scenarios: {
    crud_lifecycle: {
      executor: 'constant-vus',
      vus: VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    checks: ['rate>0.99'],
  },
};

function json(response) {
  try {
    return response.json();
  } catch (_) {
    return {};
  }
}

function request(method, path, token, body, expectedStatus, label, extraHeaders = {}) {
  const params = {
    headers: {
      Authorization: `Bearer ${token}`,
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      ...extraHeaders,
    },
    tags: { lifecycle: label },
  };
  const payload = body === undefined ? null : JSON.stringify(body);
  const response = http.request(method, `${BASE_URL}${path}`, payload, params);
  const ok = check(response, {
    [`${label}: status ${expectedStatus}`]: (r) => r.status === expectedStatus,
  });
  if (!ok) {
    fail(`${label} failed: HTTP ${response.status} ${response.body}`);
  }
  return json(response);
}

function idFrom(data, label) {
  const id = data?.data?.id;
  if (!id) fail(`${label}: response did not contain data.id: ${JSON.stringify(data)}`);
  return id;
}

export function setup() {
  const response = http.post(
    `${BASE_URL}/api/auth/login`,
    JSON.stringify({ email: 'john.doe@hellodota.com', password: 'password123' }),
    { headers: { 'Content-Type': 'application/json' }, tags: { lifecycle: 'login' } },
  );
  const data = json(response);
  const token = data?.data?.access_token;
  if (response.status !== 200 || !token) {
    throw new Error(`login failed: HTTP ${response.status} ${response.body}`);
  }
  const me = request('GET', '/api/auth/me', token, undefined, 200, 'auth-me');
  return { token, userID: me?.data?.id };
}

function uniqueSuffix() {
  return `${Date.now()}-${__VU}-${__ITER}`;
}

function userLifecycle(token, suffix) {
  const email = `k6.lifecycle.${suffix}@example.com`;
  const created = request('POST', '/api/user-command/create', token, {
    firstname: 'Kilo',
    lastname: 'Lifecycle',
    email,
    password: 'password123',
    confirm_password: 'password123',
  }, 200, 'user-create');
  const id = idFrom(created, 'user-create');

  request('GET', `/api/user-query/${id}`, token, undefined, 200, 'user-read-created');
  request('POST', `/api/user-command/update/${id}`, token, {
    user_id: id,
    firstname: 'Kilo',
    lastname: 'Updated',
    email,
    password: 'password123',
    confirm_password: 'password123',
  }, 200, 'user-update');
  request('POST', `/api/user-command/trashed/${id}`, token, undefined, 200, 'user-trash');
  request('GET', `/api/user-query/trashed`, token, undefined, 200, 'user-read-trashed');
  request('POST', `/api/user-command/restore/${id}`, token, undefined, 200, 'user-restore');
  request('GET', `/api/user-query/${id}`, token, undefined, 200, 'user-read-restored');
  request('DELETE', `/api/user-command/permanent/${id}`, token, undefined, 200, 'user-delete-permanent');
}

function roleLifecycle(token, suffix) {
  const created = request('POST', '/api/role', token, {
    name: `k6_lifecycle_role_${suffix}`,
  }, 200, 'role-create');
  const id = idFrom(created, 'role-create');

  request('GET', `/api/role-query/${id}`, token, undefined, 200, 'role-read-created');
  request('POST', `/api/role/${id}`, token, {
    name: `k6_lifecycle_role_${suffix}_updated`,
  }, 200, 'role-update');
  request('POST', `/api/role/trashed/${id}`, token, undefined, 200, 'role-trash');
  request('GET', '/api/role-query/trashed', token, undefined, 200, 'role-read-trashed');
  request('PUT', `/api/role/restore/${id}`, token, undefined, 200, 'role-restore');
  request('GET', `/api/role-query/${id}`, token, undefined, 200, 'role-read-restored');
  request('DELETE', `/api/role/permanent/${id}`, token, undefined, 200, 'role-delete-permanent');
}

function merchantLifecycle(token, userID, suffix) {
  const created = request('POST', '/api/merchant-command/create', token, {
    name: `K6 Lifecycle Merchant ${suffix}`,
    user_id: Number(userID),
  }, 200, 'merchant-create');
  const id = idFrom(created, 'merchant-create');

  request('GET', `/api/merchant-query/${id}`, token, undefined, 200, 'merchant-read-created');
  request('POST', `/api/merchant-command/updates/${id}`, token, {
    merchant_id: id,
    name: `K6 Lifecycle Merchant ${suffix} Updated`,
    user_id: Number(userID),
    status: 'active',
  }, 200, 'merchant-update');
  request('POST', `/api/merchant-command/trashed/${id}`, token, undefined, 200, 'merchant-trash');
  request('GET', '/api/merchant-query/trashed', token, undefined, 200, 'merchant-read-trashed');
  request('POST', `/api/merchant-command/restore/${id}`, token, undefined, 200, 'merchant-restore');
  request('GET', `/api/merchant-query/${id}`, token, undefined, 200, 'merchant-read-restored');
  request('DELETE', `/api/merchant-command/permanent/${id}`, token, undefined, 200, 'merchant-delete-permanent');
}

function cardLifecycle(token, userID) {
  const created = request('POST', '/api/card-command/create', token, {
    user_id: Number(userID),
    card_type: 'credit',
    expire_date: '2035-12-31T00:00:00Z',
    cvv: '321',
    card_provider: 'visa',
  }, 200, 'card-create');
  const id = idFrom(created, 'card-create');
  const cardNumber = created?.data?.card_number;

  request('GET', `/api/card-query/${id}`, token, undefined, 200, 'card-read-created');
  request('POST', `/api/card-command/update/${id}`, token, {
    card_id: id,
    user_id: Number(userID),
    card_type: 'debit',
    expire_date: '2036-12-31T00:00:00Z',
    cvv: '654',
    card_provider: 'mastercard',
  }, 200, 'card-update');
  request('POST', '/api/card-command/toggle-status', token, { card_id: id }, 200, 'card-toggle-status');
  request('POST', `/api/card-command/update-credit-limit/${id}`, token, {
    card_id: id,
    credit_limit: 2000000,
  }, 200, 'card-update-credit-limit');
  request('POST', `/api/card-command/trashed/${id}`, token, undefined, 200, 'card-trash');
  request('GET', '/api/card-query/trashed', token, undefined, 200, 'card-read-trashed');
  request('POST', `/api/card-command/restore/${id}`, token, undefined, 200, 'card-restore');
  request('GET', `/api/card-query/${id}`, token, undefined, 200, 'card-read-restored');
  request('DELETE', `/api/card-command/permanent/${id}`, token, undefined, 200, 'card-delete-permanent');
  return cardNumber;
}

export default function (data) {
  const suffix = uniqueSuffix();
  userLifecycle(data.token, suffix);
  roleLifecycle(data.token, suffix);
  merchantLifecycle(data.token, data.userID, suffix);
  cardLifecycle(data.token, data.userID);
}
