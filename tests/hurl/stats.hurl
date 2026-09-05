# Stats Transaction
GET {{host}}/api/stats/transaction/amount/monthly?year=2024
Authorization: Bearer {{token}}
HTTP 200
[Asserts]
jsonpath "$.status" == "success"

GET {{host}}/api/stats/transaction/status/success/monthly?year=2024&month=1
Authorization: Bearer {{token}}
HTTP 200

# Stats Card
GET {{host}}/api/stats/card/dashboard
Authorization: Bearer {{token}}
HTTP 200

# Stats Merchant
GET {{host}}/api/stats/merchant/transactions
Authorization: Bearer {{token}}
HTTP 200

# Stats Topup
GET {{host}}/api/stats/topup/amount/monthly?year=2024
Authorization: Bearer {{token}}
HTTP 200

# Stats Saldo
GET {{host}}/api/stats/saldo/total/monthly?year=2024
Authorization: Bearer {{token}}
HTTP 200

# Stats Transfer
GET {{host}}/api/stats/transfer/amount/monthly?year=2024
Authorization: Bearer {{token}}
HTTP 200

# Stats Withdraw
GET {{host}}/api/stats/withdraw/amount/monthly?year=2024
Authorization: Bearer {{token}}
HTTP 200
