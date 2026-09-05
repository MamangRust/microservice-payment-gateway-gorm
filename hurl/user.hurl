# User Service API Tests (microservice routes)
# Base URL: http://localhost:5000

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a new user through the user command API
POST http://localhost:5000/api/user-command/create
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "firstname": "Jane",
  "lastname": "Smith",
  "email": "jane.smith.{{unique_suffix}}@example.com",
  "password": "password123",
  "confirm_password": "password123"
}

HTTP 200
[Captures]
created_user_id: jsonpath "$.data.id"

### Get all users
GET http://localhost:5000/api/user-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get user by ID
GET http://localhost:5000/api/user-query/{{created_user_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Update the user
POST http://localhost:5000/api/user-command/update/{{created_user_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "user_id": {{created_user_id}},
  "firstname": "Jane",
  "lastname": "Smithson",
  "email": "jane.smith.{{unique_suffix}}@example.com",
  "password": "password123",
  "confirm_password": "password123"
}

HTTP 200

### Get active users
GET http://localhost:5000/api/user-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed users
GET http://localhost:5000/api/user-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Trash the user (soft delete)
POST http://localhost:5000/api/user-command/trashed/{{created_user_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore the user
POST http://localhost:5000/api/user-command/restore/{{created_user_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete the user permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/user-command/permanent/{{created_user_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
