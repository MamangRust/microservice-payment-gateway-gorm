# Withdraw Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a card to attach the withdrawal to
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
card_number: jsonpath "$.data.card_number"

### Create a saldo record for the card (withdraw needs an existing balance record)
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "total_balance": 1000000
}

HTTP 200

### Create a new withdrawal (idempotent: same key + payload replays the original result)
POST http://localhost:5000/api/withdraw-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}
Idempotency-Key: withdraw-hurl-{{card_number}}

{
  "card_number": "{{card_number}}",
  "withdraw_amount": 100000,
  "withdraw_time": "2026-08-01T10:00:00Z"
}

HTTP 200
[Captures]
withdraw_id: jsonpath "$.data.id"

### Get all withdrawals
GET http://localhost:5000/api/withdraw-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Update the withdrawal
POST http://localhost:5000/api/withdraw-command/update/{{withdraw_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "withdraw_id": {{withdraw_id}},
  "withdraw_amount": 150000,
  "withdraw_time": "2026-08-02T10:00:00Z"
}

HTTP 200

### Get withdrawal by ID
GET http://localhost:5000/api/withdraw-query/{{withdraw_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get withdrawals by card number
GET http://localhost:5000/api/withdraw-query/card-number/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active withdrawals
GET http://localhost:5000/api/withdraw-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed withdrawals
GET http://localhost:5000/api/withdraw-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Withdraw stats: amount (monthly/yearly, global)
GET http://localhost:5000/api/withdraw/stats/amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Withdraw stats: amount (monthly/yearly, by card number)
GET http://localhost:5000/api/withdraw/stats/amount/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/amount/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Withdraw stats: status success (monthly/yearly, global)
GET http://localhost:5000/api/withdraw/stats/status/monthly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/status/yearly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Withdraw stats: status failed (monthly/yearly, global)
GET http://localhost:5000/api/withdraw/stats/status/monthly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/status/yearly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Withdraw stats: status success (monthly/yearly, by card number)
GET http://localhost:5000/api/withdraw/stats/status/monthly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/status/yearly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Withdraw stats: status failed (monthly/yearly, by card number)
GET http://localhost:5000/api/withdraw/stats/status/monthly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/withdraw/stats/status/yearly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Trash the withdrawal (soft delete)
POST http://localhost:5000/api/withdraw-command/trashed/{{withdraw_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the withdrawal
POST http://localhost:5000/api/withdraw-command/restore/{{withdraw_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the withdrawal permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/withdraw-command/permanent/{{withdraw_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
