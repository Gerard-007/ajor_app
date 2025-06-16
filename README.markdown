# AjoR App API Testing Guide

This document provides instructions for testing the AjoR App API endpoints using `curl` or tools like Postman. The API is built with Go, Gin, MongoDB, and JWT authentication, supporting user registration, login, profile management, and admin functionalities.

## Prerequisites

1. **Go**: Install Go (version 1.16 or later) from [golang.org](https://golang.org).
2. **MongoDB**: Set up a MongoDB instance (local or cloud, e.g., MongoDB Atlas).
3. **Environment Variables**: Create a `.env` file in the project root with:
   ```env
   MONGODB_URI=mongodb://localhost:27017 # or your MongoDB Atlas URI
   JWT_SECRET=your-secure-secret-key # At least 32 characters
   PORT=8080 # Optional, defaults to 8080
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
   go run main.go
   ```
   The server runs on `http://localhost:8080` (or the port specified in `.env`).

3. Verify MongoDB connection:
   - Check the console for `Connected to MongoDB!`.
   - Ensure the `ajor_app_db` database is created with `users`, `profiles`, and (if using blacklisting) `blacklisted_tokens` collections.

## Testing Endpoints

All endpoints are hosted at `http://localhost:8080`. Authenticated endpoints require a JWT token in the `Authorization` header as `Bearer <token>`. Admin-only actions require a user with `is_admin: true`.

### 1. Register a User (`POST /register`)

Creates a user and their profile.

**Request**:
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "securepassword123",
    "username": "user1",
    "phone": "1234567890"
  }'
```

**Expected Response**:
- **201 Created**:
  ```json
  {"message": "User registered successfully"}
  ```
- **400 Bad Request** (e.g., duplicate email):
  ```json
  {"error": "email already exists"}
  ```

**Notes**:
- Creates a `User` in the `users` collection and a `Profile` in the `profiles` collection.
- `verified` and `is_admin` default to `false`.
- To create an admin, manually set `is_admin: true` in MongoDB for a user (e.g., using MongoDB Compass).

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
- **400 Bad Request** (wrong credentials):
  ```json
  {"error": "invalid credentials"}
  ```
- **404 Not Found** (user not found):
  ```json
  {"error": "user not found"}
  ```

**Notes**:
- Save the `<jwt_token>` for authenticated requests.
- Token expires after 72 hours (configurable in `utils/jwt.go`).

### 3. Logout (`POST /logout`)

Invalidates the session (client-side or server-side with blacklisting).

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
- **Client-Side**: Remove the token from client storage (e.g., `localStorage.removeItem('token')` in JavaScript).
- **Server-Side** (if blacklisting enabled): The token is added to `blacklisted_tokens` and cannot be used again.
- Test blacklisting by attempting an authenticated request with the blacklisted token (should return `401 Unauthorized`).

### 4. Get User by ID (`GET /users/:id`)

Retrieves a user’s details (authenticated).

**Request**:
```bash
curl -X GET http://localhost:8080/users/<user_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Example** (replace `<user_id>` with a valid ObjectID, e.g., `60c72b2f9b1d4b3c7c8d9e0f`):
```bash
curl -X GET http://localhost:8080/users/60c72b2f9b1d4b3c7c8d9e0f \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "id": "60c72b2f9b1d4b3c7c8d9e0f",
    "email": "user1@example.com",
    "username": "user1",
    "phone": "1234567890",
    "verified": false,
    "is_admin": false,
    "created_at": "2025-05-26T09:00:00Z",
    "updated_at": "2025-05-26T09:00:00Z"
  }
  ```
- **400 Bad Request** (invalid ID):
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized** (missing/invalid token):
  ```json
  {"error": "Invalid or expired token"}
  ```
- **404 Not Found**:
  ```json
  {"error": "User not found"}
  ```

**Notes**:
- Any authenticated user can access this endpoint.
- `password` is excluded from the response.

### 5. Get User Profile (`GET /profile/:id`)

Retrieves a user’s profile (assumed endpoint).

**Request**:
```bash
curl -X GET http://localhost:8080/profile/<user_id> \
  -H "Authorization: Bearer <jwt_token>"
```

**Example**:
```bash
curl -X GET http://localhost:8080/profile/60c72b2f9b1d4b3c7c8d9e0f \
  -H "Authorization: Bearer <jwt_token>"
```

**Expected Response**:
- **200 OK**:
  ```json
  {
    "id": "60c72b2f9b1d4b3c7c8d9e0g",
    "user_id": "60c72b2f9b1d4b3c7c8d9e0f",
    "bio": "",
    "location": "",
    "profile_pic": "",
    "created_at": "2025-05-26T09:00:00Z",
    "updated_at": "2025-05-26T09:00:00Z"
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

**Notes**:
- Add this route to `routes.go` if not present:
  ```go
  authenticated.GET("/profile/:id", handlers.GetUserProfileHandler(db))
  ```
- Any authenticated user can access this endpoint.

### 6. Update User Profile (`PUT /profile/:id`)

Updates a user’s profile. Non-admins can only update their own profile; admins can update any profile.

**Request**:
```bash
curl -X PUT http://localhost:8080/profile/<user_id> \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Software developer",
    "location": "New York",
    "profile_pic": "https://example.com/avatar.png"
  }'
```

**Example**:
```bash
curl -X PUT http://localhost:8080/profile/60c72b2f9b1d4b3c7c8d9e0f \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "bio": "Software developer",
    "location": "New York",
    "profile_pic": "https://example.com/avatar.png"
  }'
```

**Expected Response**:
- **200 OK**:
  ```json
  {"message": "Profile updated successfully"}
  ```
- **400 Bad Request** (invalid ID or body):
  ```json
  {"error": "Invalid user ID"}
  ```
- **401 Unauthorized**:
  ```json
  {"error": "Invalid or expired token"}
  ```
- **403 Forbidden** (non-admin updating another user’s profile):
  ```json
  {"error": "Unauthorized to update this profile"}
  ```

**Notes**:
- **Non-Admin**: Use the token of the user whose profile matches `<user_id>`.
- **Admin**: Use an admin’s token to update any profile. Create an admin user by setting `is_admin: true` in MongoDB (e.g., `db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})`).
- Verify updates in the `profiles` collection.

### 7. Delete User (`DELETE /users/:id`)

Deletes a user and their profile (admin only).

**Request**:
```bash
curl -X DELETE http://localhost:8080/users/<user_id> \
  -H "Authorization: Bearer <admin_jwt_token>"
```

**Example**:
```bash
curl -X DELETE http://localhost:8080/users/60c72b2f9b1d4b3c7c8d9e0f \
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
- **404 Not Found**:
  ```json
  {"error": "User not found"}
  ```

**Notes**:
- Requires an admin token (`is_admin: true`).
- Deletes both the user (`users` collection) and profile (`profiles` collection) atomically.
- Verify deletion in MongoDB.

## Testing Workflow

1. **Setup**:
   - Start the server (`go run main.go`).
   - Ensure MongoDB is running and `.env` is configured.

2. **Create Users**:
   - Register a regular user (`POST /register`).
   - Register an admin user and set `is_admin: true` in MongoDB:
     ```javascript
     db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})
     ```

3. **Test Authentication**:
   - Log in as a regular user and admin (`POST /login`).
   - Save their tokens.

4. **Test User and Profile Endpoints**:
   - Get user details (`GET /users/:id`) with either token.
   - Get profile (`GET /profile/:id`).
   - Update profile (`PUT /profile/:id`):
     - As the user (should succeed for own profile).
     - As another non-admin (should fail with 403).
     - As an admin (should succeed for any profile).

5. **Test Deletion**:
   - Delete a user as an admin (`DELETE /users/:id`, should succeed).
   - Attempt deletion as a non-admin (should fail with 403).

6. **Test Logout**:
   - Log out (`POST /logout`).
   - Try using the token again (should fail if blacklisting is enabled).

## Using Postman

1. Import the following collection (save as `ajor_app.postman_collection.json`):
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
             "raw": "{\"email\": \"user1@example.com\", \"password\": \"securepassword123\", \"username\": \"user1\", \"phone\": \"1234567890\"}"
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
             "raw": "{\"bio\": \"Software developer\", \"location\": \"New York\", \"profile_pic\": \"https://example.com/avatar.png\"}"
           },
           "url": "{{base_url}}/profile/{{user_id}}"
         }
       },
       {
         "name": "Delete User",
         "request": {
           "method": "DELETE",
           "header": [{"key": "Authorization", "value": "Bearer {{token}}"}],
           "url": "{{base_url}}/users/{{user_id}}"
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
  - Ensure `MONGODB_URI` is correct.
  - Check MongoDB is running (`mongod` or Atlas status).

- **JWT Errors**:
  - Verify `JWT_SECRET` is set and matches across requests.
  - Check token expiration (72 hours).

- **Admin Actions**:
  - Ensure the user has `is_admin: true` in the `users` collection.
  - Use MongoDB CLI or Compass to update:
    ```javascript
    db.users.updateOne({"email": "admin@example.com"}, {"$set": {"is_admin": true}})
    ```

- **Blacklisting**:
  - If `POST /logout` doesn’t invalidate tokens, ensure `blacklisted_tokens` collection exists and `AuthMiddleware` checks it.
  - Clean expired tokens periodically:
    ```javascript
    db.blacklisted_tokens.deleteMany({"expires_at": {"$lt": ISODate()}})
    ```

## Notes

- **ObjectIDs**: Replace `<user_id>` with valid MongoDB ObjectIDs from the `users` collection (viewable in MongoDB Compass or CLI).
- **Blacklisting**: Logout uses token blacklisting if enabled. For client-side logout, remove the token from client storage.
- **Security**: Ensure `JWT_SECRET` is secure and not committed to version control.
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
  db.Collection("blacklisted_tokens").Indexes().CreateOne(ctx, mongo.IndexModel{
      Keys: bson.M{"token": 1},
      Options: options.Index().SetUnique(true),
  })
  ```

For further assistance, check server logs or contact the developer.



ajor_app/
├── cmd/
│   └── server
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
│   │   ├── user.go
│   │   ├── wallet.go
│   │   ├── contribution.go
│   │   ├── approval.go
│   │   ├── collection.go
│   │   ├── notification.go
│   │   ├── profile.go
│   │   └── transaction.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── wallet_repository.go
│   │   ├── contribution_repository.go
│   │   ├── transaction_repository.go
│   │   ├── notification_repository.go
│   │   ├── collection_repository.go
│   │   ├── approval_repository.go
│   │   ├── profile_repository.go
│   │   └── blacklist_repository.go
│   ├── services/
│   │   ├── user_service.go
│   │   ├── contribution_service.go
│   │   ├── collection_service.go
│   │   ├── transaction_service.go
│   │   ├── notification_service.go
│   │   └── approval_service.go
│   ├── routes/
│   │   └── routes.go
│   └── utils/
│       └── jwt.go
├── pkg/
│   ├── helpers/
│   ├── jobs/
│   │    └──jobs.go
│   ├── payment/
│   │    └──flutterwave.go
│   │    └──gateway.go
│   └── utils/
│       └──jwt.go
├── tests/
│   └── (all tests here)
└── .env