# Saldo Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a card to attach the saldo record to
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

### Create a new saldo record
POST http://localhost:5000/api/saldo-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "card_number": "{{card_number}}",
  "total_balance": 500000
}

HTTP 200
[Captures]
saldo_id: jsonpath "$.data.id"

### Update the saldo record
POST http://localhost:5000/api/saldo-command/update/{{saldo_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "saldo_id": {{saldo_id}},
  "card_number": "{{card_number}}",
  "total_balance": 750000
}

HTTP 200

### Get all saldo records
GET http://localhost:5000/api/saldo-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get saldo by ID
GET http://localhost:5000/api/saldo-query/{{saldo_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get saldo by card number
GET http://localhost:5000/api/saldo-query/card_number/{{card_number}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active saldo records
GET http://localhost:5000/api/saldo-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed saldo records
GET http://localhost:5000/api/saldo-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Saldo stats: balance (monthly/yearly)
GET http://localhost:5000/api/saldo/stats/balance/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/saldo/stats/balance/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Saldo stats: total balance (monthly/yearly)
GET http://localhost:5000/api/saldo/stats/total/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/saldo/stats/total/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Saldo: list reconciliation queue
GET http://localhost:5000/api/saldo-query/reconciliation?limit=10&status=
Authorization: Bearer {{auth_token}}

HTTP 200

### Saldo: list ledger entries for the card (adjustment writes immutable ledger rows)
GET http://localhost:5000/api/saldo-query/ledger/{{card_number}}?limit=10
Authorization: Bearer {{auth_token}}

HTTP 200

### Saldo: resolve a reconciliation queue item
# The reconciler worker only enqueues mismatches on its 5-minute cycle, so a
# queue_id that cannot exist is used here; the endpoint must still return the
# expected not-found status instead of a 500.
POST http://localhost:5000/api/saldo-command/reconciliation/1/resolve
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "operation_id": "hurl-resolve-reconciliation",
  "note": "test resolve"
}

HTTP 404

### Trash the saldo record (soft delete)
POST http://localhost:5000/api/saldo-command/trashed/{{saldo_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the saldo record
POST http://localhost:5000/api/saldo-command/restore/{{saldo_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the saldo record permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/saldo-command/permanent/{{saldo_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
