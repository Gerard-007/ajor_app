# AjoR App API Testing Guide

This document provides instructions for testing the AjoR App API endpoints using `curl` or tools like Postman. The API is built with Go, Gin, MongoDB, and JWT authentication, supporting user registration, login, profile management, contributions, wallets, and admin functionalities.

## Prerequisites

1. **Go**: Install Go (version 1.16 or later) from [golang.org](https://golang.org).
2. **MongoDB**: Set up a MongoDB instance (local or cloud, e.g., MongoDB Atlas).
3. **Environment Variables**: Create a `.env` file in the project root with:
   ```env
   MONGODB_URI=mongodb://localhost:27017 # or your MongoDB Atlas URI
   DB_NAME=ajor_app_db
   JWT_SECRET=your-secure-secret-key # At least 32 characters
   PORT=8080 # Optional, defaults to 8080
   FLUTTERWAVE_API_KEY=FLWSECK_TEST-abcdef1234567890 # Your Flutterwave test key
   ```
4. **Dependencies**: Install Go dependencies:
   ```bash
   go mod tidy
   ```
   Required packages:
   - `github.com/gin-gonic/gin`
   - `go.mongodb.org/mongo-driver/mongo`
   - `golang.org/x/crypto/bcrypt`
   - `github.com/dgrijalva/jwt-go`
   - `github.com/joho/godotenv`
   - `github.com/stretchr/testify` (for testing)
5. **Tools**:
   - `curl` (command-line) or Postman for HTTP requests.
   - MongoDB Compass or CLI to inspect the database (optional).

## Running the Application

1. Clone the repository (if applicable):
   ```bash
   git clone <repository-url>
   cd ajor_app
   ```

2. Start the application:
   ```bash
   go run cmd/server/main.go
   ```
   The server runs on `http://localhost:8080` (or the port specified in `.env`).

3. Verify MongoDB connection:
   - Check the console for `Connected to MongoDB!`.
   - Ensure the `ajor_app_db` database is created with collections: `users`, `profiles`, `wallets`, `contributions`, `collections`, `approvals`, `notifications`, `blacklisted_tokens`.

## Running Tests

Automated tests are located in `tests/routes_test.go`. To run them:

```bash
go test ./tests -v
```

The tests cover all endpoints, mocking MongoDB and Flutterwave interactions. Ensure `github.com/stretchr/testify` is installed:

```bash
go get github.com/stretchr/testify
```

## Testing Endpoints

All endpoints are hosted at `http://localhost:8080`. Authenticated endpoints require a JWT token in the `Authorization` header as `Bearer <token>`. Admin-only actions require a user with `is_admin: true`.

### 1. Register a User (`POST /register`)

Creates a user, profile, and wallet with a virtual account.

**Request**:
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "securepassword123",
    "phone": "+2348062134747",
    "bvn": "11234567897"
  }'
```

**Expected Response**:
- **201 Created**:
  ```json
  {"message": "User registered successfully"}
  ```
- **400 Bad Request** (e.g., duplicate email, missing fields):
  ```json
  {"error": "email already exists"}
  ```

**Notes**:
- Username is auto-generated from email (e.g., `user1` or `user101` if taken).
- Creates entries in `users`, `profiles`, and `wallets` collections.
- Requires a valid Flutterwave test key in `.env`.

### 2. Login (`POST /login`)

Authenticates a user and returns a JWT token.

**Request**:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "securepassword123"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"token": "<jwt_token>"}
  ```
- **401 Unauthorized** (wrong credentials):
  ```json
  {"error": "Invalid credentials"}
  ```
- **404 Not Found** (user not found):
  ```json
  {"error": "user not found"}
  ```

**Notes**:
- Save `<jwt_token>` for authenticated requests.
- Token expires after 72 hours (configurable in `pkg/utils/jwt.go`).

### 3. Logout (`POST /logout`)

Blacklists the JWT token.

**Request**:
```bash
curl -X POST http://localhost:8080/logout \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Logged out successfully"}
  ```
- **400 Bad Request** (missing/invalid token):
  ```json
  {"error": "Invalid token"}
  ```

**Notes**:
- Adds token to `blacklisted_tokens` collection.
- Test blacklisting by reusing the token (should return `401 Unauthorized`).

### 4. Get User by ID (`GET /users/:id`)

Retrieves user details, profile, and wallet (authenticated user or admin).

**Request**:
```bash
curl -X GET http://localhost:8080/users/<user_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Example**:
```bash
curl -X GET http://localhost:8080/users/68514f461783445e603004d2 \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "_id": "68514f461783445e603004d2",
    "username": "user1",
    "email": "user1@example.com",
    "is_admin": false,
    "phone": "+2348062134747",
    "bvn": "11234567897",
    "created_at": "2025-06-17T11:19:34.946Z",
    "updated_at": "2025-06-17T11:19:34.946Z",
    "profile": {
      "_id": "68514f461783445e603004d3",
      "user_id": "68514f461783445e603004d2",
      "bio": "",
      "location": "",
      "profile_pic": "",
      "created_at": "2025-06-17T11:19:34.980Z",
      "updated_at": "2025-06-17T11:19:34.980Z"
    },
    "wallet": {
      "_id": "68514f471783445e603004d4",
      "owner_id": "68514f461783445e603004d2",
      "type": "user",
      "balance": 0,
      "virtual_account_id": "VA123",
      "virtual_account_number": "1234567890",
      "virtual_bank_name": "Test Bank",
      "created_at": "2025-06-17T11:19:35.000Z",
      "updated_at": "2025-06-17T11:19:42.237Z"
    }
  }
  ```
- **400 Bad Request** (invalid ID):
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-owner/non-admin):
  ```json
  {"error": "Unauthorized access"}
  ```
- **404 Not Found**:
  ```json
  {"error": "user not found"}
  ```

**Notes**:
- Non-admins can only access their own data.

### 5. Get All Users (`GET /admin/users`)

Lists all users (admin only).

**Request**:
```bash
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "68514f461783445e603004d2",
      "username": "user1",
      "email": "user1@example.com",
      "is_admin": false,
      "phone": "+2348062134747",
      "bvn": "11234567897",
      "created_at": "2025-06-17T11:19:34.946Z",
      "updated_at": "2025-06-17T11:19:34.946Z"
    }
  ]
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can access this endpoint"}
  ```

**Notes**:
- Create an admin user:
  ```javascript
  db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})
  ```

### 6. Get User Profile (`GET /profile/:id`)

Retrieves a user’s profile.

**Request**:
```bash
curl -X GET http://localhost:8080/profile/<user_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Example**:
```bash
curl -X GET http://localhost:8080/profile/68514f461783445e603004d2 \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "_id": "68514f461783445e603004d3",
    "user_id": "68514f461783445e603004d2",
    "bio": "",
    "location": "",
    "profile_pic": "",
    "created_at": "2025-06-17T11:19:34.980Z",
    "updated_at": "2025-06-17T11:19:34.980Z"
  }
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **404 Not Found**:
  ```json
  {"error": "Profile not found"}
  ```

### 7. Update User Profile (`PUT /profile/:id`)

Updates a user’s profile.

**Request**:
```bash
curl -X PUT http://localhost:8080/profile/<user_id> \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Software developer",
    "location": "Lagos",
    "profile_pic": "/uploads/profile-pic.jpg"
  }'
```

**Example**:
```bash
curl -X PUT http://localhost:8080/profile/68514f461783445e603004d2 \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Software developer",
    "location": "Lagos",
    "profile_pic": "/uploads/profile-pic.jpg"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Profile updated successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-owner/non-admin):
  ```json
  {"error": "Unauthorized to update this profile"}
  ```

### 8. Update User (`PUT /users/:id`)

Updates user details (admin only).

**Request**:
```bash
curl -X PUT http://localhost:8080/users/<user_id> \
  -H "Authorization: Bearer <admin_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "updated@example.com",
    "phone": "+2348062134748"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "User updated successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can update users"}
  ```

### 9. Delete User (`DELETE /users/:id`)

Deletes a user, profile, and wallet (admin only).

**Request**:
```bash
curl -X DELETE http://localhost:8080/users/<user_id> \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Example**:
```bash
curl -X DELETE http://localhost:8080/users/68514f461783445e603004d2 \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "User and profile deleted successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can delete users"}
  ```

### 10. Create Contribution (`POST /contributions`)

Creates a contribution group.

**Request**:
```bash
curl -X POST http://localhost:8080/contributions \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Savings Group",
    "description": "Monthly savings",
    "amount": 1000,
    "cycle": "monthly"
  }'
```

**Expected Response**:
- **201 Created**:
  ```json
  {"message": "Contribution created successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid input"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 11. Get Contribution (`GET /contributions/:id`)

Retrieves a contribution’s details.

**Request**:
```bash
curl -X GET http://localhost:8080/contributions/<contribution_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "_id": "<contribution_id>",
    "name": "Savings Group",
    "description": "Monthly savings",
    "amount": 1000,
    "cycle": "monthly",
    "creator_id": "<user_id>",
    "created_at": "2025-06-17T11:19:34.946Z"
  }
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **404 Not Found**:
  ```json
  {"error": "Contribution not found"}
  ```

### 12. Get User Contributions (`GET /contributions`)

Lists all contributions a user is part of.

**Request**:
```bash
curl -X GET http://localhost:8080/contributions \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<contribution_id>",
      "name": "Savings Group",
      "amount": 1000,
      "cycle": "monthly"
    }
  ]
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 13. Update Contribution (`PUT /contributions/:id`)

Updates a contribution’s details (creator or admin only).

**Request**:
```bash
curl -X PUT http://localhost:8080/contributions/<contribution_id> \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Savings Group",
    "amount": 1500
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Contribution updated successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-creator/non-admin):
  ```json
  {"error": "Unauthorized to update this contribution"}
  ```

### 14. Join Contribution (`POST /contributions/join`)

Joins a contribution group.

**Request**:
```bash
curl -X POST http://localhost:8080/contributions/join \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "contribution_id": "<contribution_id>"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Joined contribution successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 15. Remove Member from Contribution (`DELETE /contributions/:id/:user_id`)

Removes a member from a contribution (creator or admin only).

**Request**:
```bash
curl -X DELETE http://localhost:8080/contributions/<contribution_id>/<user_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Member removed successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-creator/non-admin):
  ```json
  {"error": "Unauthorized to remove this member"}
  ```

### 16. Record Contribution (`POST /contributions/:id/contribute`)

Records a contribution payment.

**Request**:
```bash
curl -X POST http://localhost:8080/contributions/<contribution_id>/contribute \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Contribution recorded successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 17. Record Payout (`POST /contributions/:id/payout`)

Records a payout from a contribution (admin or creator only).

**Request**:
```bash
curl -X POST http://localhost:8080/contributions/<contribution_id>/payout \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000,
    "recipient_id": "<user_id>"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Payout recorded successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-creator/non-admin):
  ```json
  {"error": "Unauthorized to record payout"}
  ```

### 18. Get User Transactions (`GET /contributions/:id/transactions`)

Lists transactions for a contribution.

**Request**:
```bash
curl -X GET http://localhost:8080/contributions/<contribution_id>/transactions \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<transaction_id>",
      "contribution_id": "<contribution_id>",
      "user_id": "<user_id>",
      "amount": 1000,
      "type": "contribution",
      "created_at": "2025-06-17T11:19:34.946Z"
    }
  ]
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 19. Get User Notifications (`GET /notifications`)

Lists user notifications.

**Request**:
```bash
curl -X GET http://localhost:8080/notifications \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<notification_id>",
      "user_id": "<user_id>",
      "message": "New contribution recorded",
      "created_at": "2025-06-17T11:19:34.946Z"
    }
  ]
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 20. Get All Contributions (`GET /admin/contributions`)

Lists all contributions (admin only).

**Request**:
```bash
curl -X GET http://localhost:8080/admin/contributions \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<contribution_id>",
      "name": "Savings Group",
      "amount": 1000,
      "cycle": "monthly"
    }
  ]
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can access this endpoint"}
  ```

### 21. Create Collection (`POST /contributions/:id/collections`)

Creates a collection for a contribution.

**Request**:
```bash
curl -X POST http://localhost:8080/contributions/<contribution_id>/collections \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1000,
    "due_date": "2025-07-01T00:00:00Z"
  }'
```

**Expected Response**:
- **201 Created**:
  ```json
  {"message": "Collection created successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 22. Get Collections (`GET /contributions/:id/collections`)

Lists collections for a contribution.

**Request**:
```bash
curl -X GET http://localhost:8080/contributions/<contribution_id>/collections \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<collection_id>",
      "contribution_id": "<contribution_id>",
      "amount": 1000,
      "due_date": "2025-07-01T00:00:00Z"
    }
  ]
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid contribution ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```

### 23. Approve Payout (`PUT /approvals/:approval_id`)

Approves a payout request (admin only).

**Request**:
```bash
curl -X PUT http://localhost:8080/approvals/<approval_id> \
  -H "Authorization: Bearer <admin_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Payout approved successfully"}
  ```
- **400 Bad Request**:
  ```json
  {"error": "Invalid approval ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can approve payouts"}
  ```

### 24. Get Pending Approvals (`GET /approvals`)

Lists pending payout approvals (admin only).

**Request**:
```bash
curl -X GET http://localhost:8080/approvals \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  [
    {
      "_id": "<approval_id>",
      "contribution_id": "<contribution_id>",
      "amount": 1000,
      "status": "pending"
    }
  ]
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin):
  ```json
  {"error": "Only admins can access this endpoint"}
  ```

### 25. Get User Wallet (`GET /wallet`)

Retrieves the user’s wallet.

**Request**:
```bash
curl -X GET http://localhost:8080/wallet \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "_id": "68514f471783445e603004d4",
    "owner_id": "68514f461783445e603004d2",
    "type": "user",
    "balance": 0,
    "virtual_account_id": "VA123",
    "virtual_account_number": "1234567890",
    "virtual_bank_name": "Test Bank",
    "created_at": "2025-06-17T11:19:35.000Z",
    "updated_at": "2025-06-17T11:19:42.237Z"
  }
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **404 Not Found**:
  ```json
  {"error": "Wallet not found"}
  ```

### 26. Delete Wallet (`DELETE /wallet`)

Deletes the user’s wallet.

**Request**:
```bash
curl -X DELETE http://localhost:8080/wallet \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Wallet deleted successfully"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **404 Not Found**:
  ```json
  {"error": "Wallet not found"}
  ```

## Testing Workflow

1. **Setup**:
   - Start the server (`go run cmd/server/main.go`).
   - Ensure MongoDB is running and `.env` is configured.
   - Run tests (`go test ./tests -v`).

2. **Create Users**:
   - Register a regular user (`POST /register`).
   - Register an admin user and set `is_admin: true`:
     ```javascript
     db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})
     ```

3. **Test Authentication**:
   - Log in as regular user and admin (`POST /login`).
   - Save tokens.

4. **Test Endpoints**:
   - **User/Profile**: Get user (`GET /users/:id`), profile (`GET /profile/:id`), update profile (`PUT /profile/:id`), delete user (`DELETE /users/:id` as admin).
   - **Contributions**: Create (`POST /contributions`), join (`POST /contributions/join`), contribute (`POST /contributions/:id/contribute`), payout (`POST /contributions/:id/payout`).
   - **Wallet**: Get (`GET /wallet`), delete (`DELETE /wallet`).
   - **Admin**: List users (`GET /admin/users`), contributions (`GET /admin/contributions`), approve payouts (`PUT /approvals/:approval_id`).

5. **Test Logout**:
   - Log out (`POST /logout`) and verify token invalidation.

## Using Postman

1. Import the following Postman collection (`ajor_app.postman_collection.json`):
   ```json
   {
     "info": {
       "name": "AjoR App API",
       "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
     },
     "item": [
       {
         "name": "Register",
         "request": {
           "method": "POST",
           "header": [{"key": "Content-Type", "value": "application/json"}],
           "body": {
             "mode": "raw",
             "raw": "{\"email\": \"user1@example.com\", \"password\": \"securepassword123\", \"phone\": \"+2348062134747\", \"bvn\": \"11234567897\"}"
           },
           "url": "{{base_url}}/register"
         }
       },
       {
         "name": "Login",
         "request": {
           "method": "POST",
           "header": [{"key": "Content-Type", "value": "application/json"}],
           "body": {
             "mode": "raw",
             "raw": "{\"email\": \"user1@example.com\", \"password\": \"securepassword123\"}"
           },
           "url": "{{base_url}}/login"
         }
       },
       {
         "name": "Logout",
         "request": {
           "method": "POST",
           "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
           "url": "{{base_url}}/logout"
         }
       },
       {
         "name": "Get User",
         "request": {
           "method": "GET",
           "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
           "url": "{{base_url}}/users/{{user_id}}"
         }
       },
       {
         "name": "Get Profile",
         "request": {
           "method": "GET",
           "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
           "url": "{{base_url}}/profile/{{user_id}}"
         }
       },
       {
         "name": "Update Profile",
         "request": {
           "method": "PUT",
           "header": [
             {"key": "Authorization", "value": "Bearer {{token}}"},
             {"key": "Content-Type", "value": "application/json"}
           ],
           "body": {
             "mode": "raw",
             "raw": "{\"bio\": \"Software developer\", \"location\": \"Lagos\", \"profile_pic\": \"/uploads/profile-pic.jpg\"}"
           },
           "url": "{{base_url}}/profile/{{user_id}}"
         }
       },
       {
         "name": "Get Wallet",
         "request": {
           "method": "GET",
           "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
           "url": "{{base_url}}/wallet"
         }
       },
       {
         "name": "Create Contribution",
         "request": {
           "method": "POST",
           "header": [
             {"key": "Authorization", "value": "Bearer {{token}}"},
             {"key": "Content-Type", "value": "application/json"}
           ],
           "body": {
             "mode": "raw",
             "raw": "{\"name\": \"Savings Group\", \"description\": \"Monthly savings\", \"amount\": 1000, \"cycle\": \"monthly\"}"
           },
           "url": "{{base_url}}/contributions"
         }
       }
     ],
     "variable": [
       {"key": "base_url", "value": "http://localhost:8080"},
       {"key": "token", "value": ""},
       {"key": "user_id", "value": ""}
     ]
   }
   ```
2. Set environment variables in Postman:
   - `base_url`: `http://localhost:8080`
   - `token`: Set after `POST /login`.
   - `user_id`: Set to a valid ObjectID.

3. Run requests and verify responses.

## Troubleshooting

- **MongoDB Connection**:
  - Ensure `MONGODB_URI` and `DB_NAME` are correct.
  - Check MongoDB is running (`mongod` or Atlas status).

- **JWT Errors**:
  - Verify `JWT_SECRET` is set and consistent.
  - Check token expiration (72 hours).

- **Flutterwave Errors**:
  - Ensure `FLUTTERWAVE_API_KEY` is a valid test key.
  - Check Flutterwave dashboard for test data (e.g., BVN: `11234567897`).

- **Admin Actions**:
  - Set `is_admin: true` for admin users:
    ```javascript
    db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})
    ```

- **Blacklisting**:
  - Ensure `blacklisted_tokens` collection exists.
  - Clean expired tokens:
    ```javascript
    db.blacklisted_tokens.deleteMany({"expires_at": {"$lt": ISODate()}})
    ```

## Notes

- **ObjectIDs**: Use valid MongoDB ObjectIDs from collections (viewable in MongoDB Compass or CLI).
- **Security**: Keep `JWT_SECRET` and `FLUTTERWAVE_API_KEY` secure.
- **Indexes**: Add indexes for performance (in `repository.InitDatabase`):
  ```go
  usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
      Keys: bson.M{"email": 1},
      Options: options.Index().SetUnique(true),
  })
  usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
      Keys: bson.M{"username": 1},
      Options: options.Index().SetUnique(true),
  })
  profilesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
      Keys: bson.M{"user_id": 1},
      Options: options.Index().SetUnique(true),
  })
  walletsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
      Keys: bson.M{"owner_id": 1},
      Options: options.Index().SetUnique(true),
  })
  ```

## Project Structure

```
ajor_app/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── middleware.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── wallet_handler.go
│   │   ├── contribution_handler.go
│   │   ├── collection_handler.go
│   │   ├── transaction_handler.go
│   │   ├── notification_handler.go
│   │   ├── approval_handler.go
│   │   └── profile_handler.go
│   ├── models/
│   │   └── models.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── wallet_repository.go
│   │   ├── contribution_repository.go
│   │   ├── transaction_repository.go
│   │   ├── notification_repository.go
│   │   ├── collection_repository.go
│   │   ├── approval_repository.go
│   │   ├── profile_repository.go
│   │   └── auth_repository.go
│   ├── services/
│   │   ├── user_service.go
│   │   ├── auth_service.go
│   │   ├── contribution_service.go
│   │   ├── collection_service.go
│   │   ├── transaction_service.go
│   │   ├── notification_service.go
│   │   ├── wallet_service.go
│   │   ├── approval_service.go
│   │   └── profile_service.go
│   └── routes/
│       └── routes.go
├── pkg/
│   ├── payment/
│   │   ├── flutterwave.go
│   │   └── gateway.go
│   └── utils/
│       ├── jwt.go
│       └── username.go
├── tests/
│   └── routes_test.go
└── .env
```

For further assistance, check server logs or contact the developer.