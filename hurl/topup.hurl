# Topup Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a card to attach the topup to
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

### Create a saldo record for the card (topup needs an existing balance record)
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "total_balance": 500000
}

HTTP 200

### Create a topup (idempotent: same key + payload replays the original result)
POST http://localhost:5000/api/topup-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}
Idempotency-Key: topup-hurl-{{card_number}}

{
  "card_number": "{{card_number}}",
  "topup_amount": 100000,
  "topup_method": "bri"
}

HTTP 200
[Captures]
topup_id: jsonpath "$.data.id"

### Get all topups
GET http://localhost:5000/api/topup-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get topup by ID
GET http://localhost:5000/api/topup-query/{{topup_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{topup_id}}

### Update the topup
POST http://localhost:5000/api/topup-command/update/{{topup_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "topup_id": {{topup_id}},
  "topup_amount": 150000,
  "topup_method": "bca"
}

HTTP 200

### Get topups by card number
GET http://localhost:5000/api/topup-query/card-number/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active topups
GET http://localhost:5000/api/topup-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed topups
GET http://localhost:5000/api/topup-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Topup stats: amount (monthly/yearly, global)
GET http://localhost:5000/api/topup/stats/amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: amount (monthly/yearly, by card number)
GET http://localhost:5000/api/topup/stats/amount/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/amount/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: method (monthly/yearly, global)
GET http://localhost:5000/api/topup/stats/method/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/method/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: method (monthly/yearly, by card number)
GET http://localhost:5000/api/topup/stats/method/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/method/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: status success (monthly/yearly, global)
GET http://localhost:5000/api/topup/stats/status/monthly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/status/yearly/success?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: status failed (monthly/yearly, global)
GET http://localhost:5000/api/topup/stats/status/monthly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/status/yearly/failed?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: status success (monthly/yearly, by card number)
GET http://localhost:5000/api/topup/stats/status/monthly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/status/yearly/success/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Topup stats: status failed (monthly/yearly, by card number)
GET http://localhost:5000/api/topup/stats/status/monthly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/topup/stats/status/yearly/failed/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Trash the topup (soft delete)
POST http://localhost:5000/api/topup-command/trashed/{{topup_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the topup
POST http://localhost:5000/api/topup-command/restore/{{topup_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the topup permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/topup-command/permanent/{{topup_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
