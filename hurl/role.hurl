# Role Service API Tests (microservice routes)
# Base URL: http://localhost:5000
#
# NOTE: role command + query routes are exercised with a unique role name
# (injected via {{unique_suffix}} from run_tests.sh) so the suite is repeatable.

### Resolve the authenticated user id
GET http://localhost:5000/api/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Captures]
user_id: jsonpath "$.data.id"

### Create a role
POST http://localhost:5000/api/role
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "name": "hurl_e2e_role_{{unique_suffix}}"
}

HTTP 200
[Captures]
role_id: jsonpath "$.data.id"

### Get all roles
GET http://localhost:5000/api/role-query
Authorization: Bearer {{auth_token}}

HTTP 200

### Get role by ID
GET http://localhost:5000/api/role-query/{{role_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{role_id}}

### Get active roles
GET http://localhost:5000/api/role-query/active
Authorization: Bearer {{auth_token}}

HTTP 200

### Get roles by user id
GET http://localhost:5000/api/role-query/user/{{user_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Update role
POST http://localhost:5000/api/role/{{role_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}

{
  "name": "hurl_e2e_role_{{unique_suffix}}_updated"
}

HTTP 200

### Trash role (soft delete)
POST http://localhost:5000/api/role/trashed/{{role_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Get trashed roles
GET http://localhost:5000/api/role-query/trashed
Authorization: Bearer {{auth_token}}

HTTP 200

### Restore role
PUT http://localhost:5000/api/role/restore/{{role_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

### Delete role permanently (cleanup so the test is repeatable)
DELETE http://localhost:5000/api/role/permanent/{{role_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
