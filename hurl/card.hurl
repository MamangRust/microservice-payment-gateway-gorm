# Card Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a new card
POST http://localhost:5000/api/card-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "user_id": {{user_id}},
  "card_type": "credit",
  "expire_date": "2030-12-31T00:00:00Z",
  "cvv": "123",
  "card_provider": "visa"
}

HTTP 200
[Captures]
card_id: jsonpath "$.data.id"
card_number: jsonpath "$.data.card_number"

### Update the card
POST http://localhost:5000/api/card-command/update/{{card_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_id": {{card_id}},
  "user_id": {{user_id}},
  "card_type": "debit",
  "expire_date": "2031-12-31T00:00:00Z",
  "cvv": "456",
  "card_provider": "mastercard"
}

HTTP 200

### Toggle card status
POST http://localhost:5000/api/card-command/toggle-status
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_id": {{card_id}}
}

HTTP 200

### Update credit limit
POST http://localhost:5000/api/card-command/update-credit-limit/{{card_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_id": {{card_id}},
  "credit_limit": 2000000
}

HTTP 200

### Redeem points with an insufficient balance returns 400
POST http://localhost:5000/api/card-command/redeem-points/{{card_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_id": {{card_id}},
  "points": 999999
}

HTTP 400

### Get all cards
GET http://localhost:5000/api/card-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get card by ID
GET http://localhost:5000/api/card-query/{{card_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get cards by user (user id is resolved from the JWT context)
GET http://localhost:5000/api/card-query/user
Authorization: Bearer {{auth_token}}

HTTP 200

### Get card by card number
GET http://localhost:5000/api/card-query/card_number/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active cards
GET http://localhost:5000/api/card-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed cards
GET http://localhost:5000/api/card-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Card stats: balance (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/balance/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/balance/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: balance (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/balance/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/balance/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: topup (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/topup/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/topup/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: topup (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/topup/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/topup/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transaction (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/transaction/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transaction/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transaction (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/transaction/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transaction/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transfer sender (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/transfer/sender/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transfer/sender/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transfer sender (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/transfer/sender/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transfer/sender/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transfer receiver (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/transfer/receiver/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transfer/receiver/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: transfer receiver (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/transfer/receiver/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/transfer/receiver/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: withdraw (monthly/yearly, global)
GET http://localhost:5000/api/card/stats/withdraw/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/withdraw/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: withdraw (monthly/yearly, by card number)
GET http://localhost:5000/api/card/stats/withdraw/monthly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/card/stats/withdraw/yearly/{{card_number}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: dashboard (global)
GET http://localhost:5000/api/card/stats/dashboard
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Card stats: dashboard (by card number)
GET http://localhost:5000/api/card/stats/dashboard/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Process billing cycles (generates statements for the latest closed period)
POST http://localhost:5000/api/card-command/process-billing
Authorization: Bearer {{auth_token}}

HTTP 200

### Trash the card (soft delete)
POST http://localhost:5000/api/card-command/trashed/{{card_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the card
POST http://localhost:5000/api/card-command/restore/{{card_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the card permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/card-command/permanent/{{card_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
