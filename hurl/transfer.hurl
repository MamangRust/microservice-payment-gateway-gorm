# Transfer Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create the sender card
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
sender_card_number: jsonpath "$.data.card_number"

### Create a saldo record for the sender card
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{sender_card_number}}",
  "total_balance": 1000000
}

HTTP 200

### Create the receiver card
POST http://localhost:5000/api/card-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "user_id": {{user_id}},
  "card_type": "debit",
  "expire_date": "2031-12-31T00:00:00Z",
  "cvv": "456",
  "card_provider": "mastercard"
}

HTTP 200
[Captures]
receiver_card_number: jsonpath "$.data.card_number"

### Create a saldo record for the receiver card
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{receiver_card_number}}",
  "total_balance": 1000000
}

HTTP 200

### Create a transfer (idempotent: same key + payload replays the original result)
POST http://localhost:5000/api/transfer-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}
Idempotency-Key: transfer-hurl-{{sender_card_number}}

{
  "transfer_from": "{{sender_card_number}}",
  "transfer_to": "{{receiver_card_number}}",
  "transfer_amount": 50000
}

HTTP 200
[Captures]
transfer_id: jsonpath "$.data.id"

### Get all transfers
GET http://localhost:5000/api/transfer-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get transfer by ID
GET http://localhost:5000/api/transfer-query/{{transfer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{transfer_id}}

### Update the transfer
POST http://localhost:5000/api/transfer-command/update/{{transfer_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "transfer_id": {{transfer_id}},
  "transfer_from": "{{sender_card_number}}",
  "transfer_to": "{{receiver_card_number}}",
  "transfer_amount": 60000
}

HTTP 200

### Get transfers from the sender card
GET http://localhost:5000/api/transfer-query/transfer_from/{{sender_card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get transfers to the receiver card
GET http://localhost:5000/api/transfer-query/transfer_to/{{receiver_card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active transfers
GET http://localhost:5000/api/transfer-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed transfers
GET http://localhost:5000/api/transfer-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Transfer stats: amount (monthly/yearly, global)
GET http://localhost:5000/api/transfer/stats/amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: amount by sender (monthly/yearly)
GET http://localhost:5000/api/transfer/stats/amount/monthly/sender/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/amount/yearly/sender/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: amount by receiver (monthly/yearly)
GET http://localhost:5000/api/transfer/stats/amount/monthly/receiver/{{receiver_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/amount/yearly/receiver/{{receiver_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: status success (monthly/yearly, global)
GET http://localhost:5000/api/transfer/stats/status/monthly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/status/yearly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: status failed (monthly/yearly, global)
GET http://localhost:5000/api/transfer/stats/status/monthly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/status/yearly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: status success (monthly/yearly, by sender card number)
GET http://localhost:5000/api/transfer/stats/status/monthly/success/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/status/yearly/success/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Transfer stats: status failed (monthly/yearly, by sender card number)
GET http://localhost:5000/api/transfer/stats/status/monthly/failed/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/transfer/stats/status/yearly/failed/{{sender_card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Trash the transfer (soft delete)
POST http://localhost:5000/api/transfer-command/trashed/{{transfer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the transfer
POST http://localhost:5000/api/transfer-command/restore/{{transfer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the transfer permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/transfer-command/permanent/{{transfer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
