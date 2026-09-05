# Hurl API Test Files

This directory contains Hurl test files for all API endpoints in the monolith payment gateway. Each file corresponds to a specific service and contains test cases for all available endpoints.

## Available Test Files

| Service | File | Description |
|---------|------|-------------|
| Auth | `auth.hurl` | Authentication endpoints (login, register, token refresh, etc.) |
| Card | `card.hurl` | Card management and related operations |
| Merchant | `merchant.hurl` | Merchant management operations |
| Merchant Document | `merchantdocument.hurl` | Merchant document management |
| Role | `role.hurl` | Role-based access control |
| Saldo | `saldo.hurl` | Balance and account management |
| Topup | `topup.hurl` | Top-up operations |
| Transaction | `transaction.hurl` | Transaction management |
| Transfer | `transfer.hurl` | Fund transfer operations |
| User | `user.hurl` | User management |
| Withdraw | `withdraw.hurl` | Withdrawal operations |

## Prerequisites

1. **Install Hurl**: Follow the installation guide at [https://hurl.dev/docs/installation.html](https://hurl.dev/docs/installation.html)

2. **Start the API Gateway**: Make sure the API Gateway is running on `http://localhost:8080`

## Usage

### Running Individual Tests

```bash
# Test auth endpoints
hurl auth.hurl

# Test card endpoints
hurl card.hurl

# Test merchant endpoints
hurl merchant.hurl
```

### Running All Tests

```bash
# Run all test files
hurl *.hurl

# Or run them one by one in a loop
for file in *.hurl; do
    echo "Running $file..."
    hurl "$file"
done
```

### Variables

The test files use placeholder variables that you need to replace:

- `{{auth_token}}` - Replace with a valid JWT access token
- `{{refresh_token}}` - Replace with a valid refresh token
- `{{reset_token}}` - Replace with a valid password reset token

### Getting Auth Tokens

1. First, register a user:
```bash
hurl auth.hurl --test "Register a new user"
```

2. Then login to get tokens:
```bash
hurl auth.hurl --test "Login user"
```

3. Extract the tokens from the response and update the files or use environment variables.

### Using Environment Variables

You can set environment variables instead of editing files:

```bash
export AUTH_TOKEN="your_jwt_token_here"
export REFRESH_TOKEN="your_refresh_token_here"

# Hurl will automatically use environment variables
hurl auth.hurl
```

## Test Structure

Each test file follows this structure:

1. **POST requests** - Create operations
2. **GET requests** - Read operations (all and by ID)
3. **PUT requests** - Update operations
4. **DELETE requests** - Delete operations
5. **Additional endpoints** - Service-specific endpoints

## Example Test Output

```bash
$ hurl auth.hurl
### Register a new user
POST http://localhost:8080/api/auth/register
...
HTTP/1.1 200 OK
Content-Type: application/json
...

### Login user
POST http://localhost:8080/api/auth/login
...
HTTP/1.1 200 OK
Content-Type: application/json
...
```

## Troubleshooting

### Connection Issues
- Ensure the API Gateway is running on port 8080
- Check firewall settings
- Verify the service endpoints are accessible

### Authentication Issues
- Make sure tokens are valid and not expired
- Check token format (should include "Bearer " prefix)
- Verify user permissions for specific endpoints

### Response Validation
- Hurl automatically validates HTTP status codes
- Check response bodies for error messages
- Verify request payloads match expected schemas

## Customization

You can modify the test files to:

- Change test data
- Add new test cases
- Modify expected HTTP status codes
- Add response validation
- Include additional headers

For more advanced Hurl features, see the [Hurl documentation](https://hurl.dev/docs/).