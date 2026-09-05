# Bulk Operations Coverage (runs LAST because file name sorts after the others)
#
# restore/all is safe anywhere, but permanent/all physically deletes every row
# in the table (including the shared john.doe@hellodota.com user used by later
# files in the suite), so this file must stay the lexically-last *.hurl file.
# The run_tests.sh runner iterates `*.hurl` in sorted order, hence the zzz_ prefix.

### Card: restore all
POST http://localhost:5000/api/card-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Card: permanent delete all
POST http://localhost:5000/api/card-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Merchant: restore all
POST http://localhost:5000/api/merchant-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Merchant: permanent delete all
POST http://localhost:5000/api/merchant-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Merchant Document: restore all
POST http://localhost:5000/api/merchant-document-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Merchant Document: permanent delete all
POST http://localhost:5000/api/merchant-document-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Role: restore all
POST http://localhost:5000/api/role/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Role: permanent delete all
POST http://localhost:5000/api/role/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Saldo: restore all
POST http://localhost:5000/api/saldo-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Saldo: permanent delete all
POST http://localhost:5000/api/saldo-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Topup: restore all
POST http://localhost:5000/api/topup-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Topup: permanent delete all
POST http://localhost:5000/api/topup-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Transaction: restore all
POST http://localhost:5000/api/transaction-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Transaction: permanent delete all
POST http://localhost:5000/api/transaction-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Transfer: restore all
POST http://localhost:5000/api/transfer-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Transfer: permanent delete all
POST http://localhost:5000/api/transfer-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### User: restore all
POST http://localhost:5000/api/user-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### User: permanent delete all
POST http://localhost:5000/api/user-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Withdraw: restore all
POST http://localhost:5000/api/withdraw-command/restore/all
Authorization: Bearer {{auth_token}}

HTTP 200

### Withdraw: permanent delete all
POST http://localhost:5000/api/withdraw-command/permanent/all
Authorization: Bearer {{auth_token}}

HTTP 200
