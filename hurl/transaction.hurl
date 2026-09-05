# Transaction Service API Tests (microservice routes)
# Base URL: http://localhost:5000
#
# NOTE: transaction-command/create and /update are protected by the
# ApiKeyValidator middleware, which requires a valid merchant API key in the
# X-Api-Key header. The key is captured from merchant-command/create below.

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a customer card
POST http://localhost:5000/api/card-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "user_id": {{user_id}},
  "card_type": "debit",
  "expire_date": "2030-12-31T00:00:00Z",
  "cvv": "123",
  "card_provider": "visa"
}

HTTP 200
[Captures]
card_id: jsonpath "$.data.id"
card_number: jsonpath "$.data.card_number"

### Create a saldo record for the customer card
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "total_balance": 1
}

HTTP 200

### Seed customer funds through the append-only balance ledger
# Card creation asynchronously initializes a zero-balance saldo. CreateSaldo is
# intentionally idempotent and preserves that existing row, so test funds must
# be applied as an explicit ledger adjustment rather than an absolute rewrite.
POST http://localhost:5000/api/saldo-command/adjustment
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "delta": 999999,
  "operation_id": "hurl-transaction-seed-{{card_number}}",
  "source_type": "test_seed",
  "source_id": "transaction-hurl",
  "note": "Seed customer balance for transaction API test"
}

HTTP 200

### Create a merchant (its API key authorizes the transaction)
POST http://localhost:5000/api/merchant-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "name": "Transaction Merchant",
  "user_id": {{user_id}}
}

HTTP 200
[Captures]
merchant_id: jsonpath "$.data.id"
merchant_api_key: jsonpath "$.data.api_key"

### Create the merchant receiving card (the merchant user's card is used as the payout target)
POST http://localhost:5000/api/card-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "user_id": {{user_id}},
  "card_type": "credit",
  "expire_date": "2031-12-31T00:00:00Z",
  "cvv": "456",
  "card_provider": "mastercard"
}

HTTP 200
[Captures]
merchant_card_number: jsonpath "$.data.card_number"

### Create a saldo record for the merchant receiving card
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{merchant_card_number}}",
  "total_balance": 1000000
}

HTTP 200

### Create a transaction (idempotent: same key + payload replays the original result)
POST http://localhost:5000/api/transaction-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}
X-Api-Key: {{merchant_api_key}}
Idempotency-Key: transaction-hurl-{{card_number}}

{
  "card_number": "{{card_number}}",
  "amount": 50000,
  "payment_method": "visa",
  "merchant_id": {{merchant_id}},
  "transaction_time": "2026-08-01T10:00:00Z"
}

HTTP 200
[Captures]
transaction_id: jsonpath "$.data.id"

### Get all transactions
GET http://localhost:5000/api/transaction-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get transaction by ID
GET http://localhost:5000/api/transaction-query/{{transaction_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{transaction_id}}

### Update the transaction (protected by the merchant API key middleware)
POST http://localhost:5000/api/transaction-command/update/{{transaction_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}
X-Api-Key: {{merchant_api_key}}

{
  "transaction_id": {{transaction_id}},
  "card_number": "{{card_number}}",
  "amount": 60000,
  "payment_method": "mastercard",
  "merchant_id": {{merchant_id}},
  "transaction_time": "2026-08-02T10:00:00Z"
}

HTTP 200

### Get transactions by card number
GET http://localhost:5000/api/transaction-query/card-number/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get transactions by merchant
GET http://localhost:5000/api/transaction-query/merchant/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active transactions
GET http://localhost:5000/api/transaction-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed transactions
GET http://localhost:5000/api/transaction-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Transaction stats: amount (monthly/yearly, global)
GET http://localhost:5000/api/transaction/stats/amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: amount (monthly/yearly, by card number)
GET http://localhost:5000/api/transaction/stats/amount/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/amount/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: method (monthly/yearly, global)
GET http://localhost:5000/api/transaction/stats/method/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/method/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: method (monthly/yearly, by card number)
GET http://localhost:5000/api/transaction/stats/method/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/method/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: status success (monthly/yearly, global)
GET http://localhost:5000/api/transaction/stats/status/monthly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/status/yearly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: status failed (monthly/yearly, global)
GET http://localhost:5000/api/transaction/stats/status/monthly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/status/yearly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: status success (monthly/yearly, by card number)
GET http://localhost:5000/api/transaction/stats/status/monthly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/status/yearly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transaction stats: status failed (monthly/yearly, by card number)
GET http://localhost:5000/api/transaction/stats/status/monthly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transaction/stats/status/yearly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Trash the transaction (soft delete)
POST http://localhost:5000/api/transaction-command/trashed/{{transaction_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the transaction
POST http://localhost:5000/api/transaction-command/restore/{{transaction_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the transaction permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/transaction-command/permanent/{{transaction_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
