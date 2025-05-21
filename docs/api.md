# Auth Service API

## Endpoints

### Authentication

| Endpoint            | Method | Description       | Request Body                                  | Response                                          |
| ------------------- | ------ | ----------------- | --------------------------------------------- | ------------------------------------------------- |
| `/register`         | POST   | Register new user | `{ "email": "string", "password": "string" }` | `{ "message": "User {{user}} created" }`          |
| `/login`            | POST   | Authenticate user | `{ "email": "string", "password": "string" }` | `{ "message": "login success" }` + session cookie |
| `/logout`           | POST   | End a session     | `{}` (requires cookie)                        | `{ "message": "logged out successfully" }`        |
| `/logouteverywhere` | POST   | End all sessions  | `{}` (requires cookie)                        | `{ "message": "logged out everywhere" }`          |

### User Management

| Endpoint         | Method | Description                  | Request Body                                                                   | Response                                                                               |
| ---------------- | ------ | ---------------------------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------- |
| `/whoami`        | GET    | Get current user information | `{}` (requires cookie)                                                         | `{ "clientIP": "string", "email": "string", "lastLogin": "date", "userID": "string" }` |
| `/updateuser`    | POST   | Update user details          | `{ "email": "string", "password": "string" }` (both optional, requires cookie) | `{ "message": "user updated" }`                                                        |
| `/deleteaccount` | POST   | Delete user account          | `{}` (requires cookie)                                                         | `{ "message": "account deleted" }`                                                     |

## Error Handling

- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Authentication required or invalid credentials
- `500 Internal Server Error`: Server error during processing

## Authentication

New sessions are stored on the client side as cookies with an expiration time and checked against a corresponding session in the database. Logout invalidates the session.
