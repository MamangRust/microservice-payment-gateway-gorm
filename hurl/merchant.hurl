# Merchant Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a new merchant
POST http://localhost:5000/api/merchant-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "name": "Toko Sejahtera",
  "user_id": {{user_id}}
}

HTTP 200
[Captures]
merchant_id: jsonpath "$.data.id"
api_key: jsonpath "$.data.api_key"

### Update the merchant
POST http://localhost:5000/api/merchant-command/updates/{{merchant_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "merchant_id": {{merchant_id}},
  "name": "Toko Sejahtera Baru",
  "user_id": {{user_id}},
  "status": "active"
}

HTTP 200

### Get all merchants
GET http://localhost:5000/api/merchant-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get merchant by ID
GET http://localhost:5000/api/merchant-query/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get merchant by API key
GET http://localhost:5000/api/merchant-query/api-key?api_key={{api_key}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active merchants
GET http://localhost:5000/api/merchant-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed merchants
GET http://localhost:5000/api/merchant-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Get merchant by user id (user id resolved from the JWT context)
GET http://localhost:5000/api/merchant-query/merchant-user
Authorization: Bearer {{auth_token}}

HTTP 200

### Merchant stats: transactions (global, by id, by apikey)
GET http://localhost:5000/api/merchant/stats/transactions
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/transactions/id/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/transactions/apikey?apikey={{api_key}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: amount (monthly/yearly, global)
GET http://localhost:5000/api/merchant/stats/amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: amount (monthly/yearly, by merchant id)
GET http://localhost:5000/api/merchant/stats/amount/monthly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/amount/yearly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: amount (monthly/yearly, by apikey)
GET http://localhost:5000/api/merchant/stats/amount/monthly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/amount/yearly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: total amount (monthly/yearly, global)
GET http://localhost:5000/api/merchant/stats/total-amount/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/total-amount/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: total amount (monthly/yearly, by merchant id)
GET http://localhost:5000/api/merchant/stats/total-amount/monthly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/total-amount/yearly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: total amount (monthly/yearly, by apikey)
GET http://localhost:5000/api/merchant/stats/total-amount/monthly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/total-amount/yearly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: method (monthly/yearly, global)
GET http://localhost:5000/api/merchant/stats/method/monthly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/method/yearly?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: method (monthly/yearly, by merchant id)
GET http://localhost:5000/api/merchant/stats/method/monthly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/method/yearly/id/{{merchant_id}}?year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Merchant stats: method (monthly/yearly, by apikey)
GET http://localhost:5000/api/merchant/stats/method/monthly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET http://localhost:5000/api/merchant/stats/method/yearly/apikey?apikey={{api_key}}&year={{year}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "success"

### Trash the merchant (soft delete)
POST http://localhost:5000/api/merchant-command/trashed/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the merchant
POST http://localhost:5000/api/merchant-command/restore/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the merchant permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/merchant-command/permanent/{{merchant_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
