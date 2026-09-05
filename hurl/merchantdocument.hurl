# Merchant Document Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a merchant to attach the document to
POST http://localhost:5000/api/merchant-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "name": "Document Merchant",
  "user_id": {{user_id}}
}

HTTP 200
[Captures]
merchant_id: jsonpath "$.data.id"

### Create a merchant document
POST http://localhost:5000/api/merchant-document-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "merchant_id": {{merchant_id}},
  "document_type": "license",
  "document_url": "https://example.com/license.pdf"
}

HTTP 200
[Captures]
document_id: jsonpath "$.data.id"

### Update the merchant document
POST http://localhost:5000/api/merchant-document-command/updates/{{document_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "document_id": {{document_id}},
  "merchant_id": {{merchant_id}},
  "document_type": "identity_proof",
  "document_url": "https://example.com/id.pdf",
  "status": "pending",
  "note": "resubmitted"
}

HTTP 200

### Update the document review status
POST http://localhost:5000/api/merchant-document-command/update-status/{{document_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "document_id": {{document_id}},
  "merchant_id": {{merchant_id}},
  "status": "approved",
  "note": "verified"
}

HTTP 200

### Get all merchant documents
GET http://localhost:5000/api/merchant-document-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get merchant document by ID
GET http://localhost:5000/api/merchant-document-query/{{document_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get active merchant documents
GET http://localhost:5000/api/merchant-document-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed merchant documents
GET http://localhost:5000/api/merchant-document-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Trash the document (soft delete)
POST http://localhost:5000/api/merchant-document-command/trashed/{{document_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the document
POST http://localhost:5000/api/merchant-document-command/restore/{{document_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the document permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/merchant-document-command/permanent/{{document_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
