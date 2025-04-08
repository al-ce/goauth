# Auth Service API

Base URL: `https://localhost:3001`

## Endpoints

### Authentication

| Endpoint            | Method | Description       | Request Body                                  | Response                                      |
| ------------------- | ------ | ----------------- | --------------------------------------------- | --------------------------------------------- |
| `/register`         | POST   | Register new user | `{ "email": "string", "password": "string" }` | `{ "message": "User {{user}} created" }`      |
| `/login`            | POST   | Authenticate user | `{ "email": "string", "password": "string" }` | `{ "message": "login success" }` + JWT Cookie |
| `/logout`           | POST   | End a session     | `{}` (requires cookie)                        | `{ "message": "logged out successfully" }`    |
| `/logouteverywhere` | POST   | End all sessions  | `{}` (requires cookie)                        | `{ "message": "logged out everywhere" }`      |

### User Management

| Endpoint         | Method | Description         | Request Body                                                                   | Response                                     |
| ---------------- | ------ | ------------------- | ------------------------------------------------------------------------------ | -------------------------------------------- |
| `/profile`       | GET    | Get user profile    | `{}` (requires cookie)                                                         | `{ "email": "string", "lastLogin": "date" }` |
| `/updateuser`    | POST   | Update user details | `{ "email": "string", "password": "string" }` (both optional, requires cookie) | `{ "message": "user updated" }`              |
| `/deleteaccount` | POST   | Delete user account | `{}` (requires cookie)                                                         | `{ "message": "account deleted" }`           |

## Error Handling

- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Authentication required or invalid credentials
- `500 Internal Server Error`: Server error during processing

## Authentication

The service uses JWT (JSON Web Tokens) for authentication. On successful login, a JWT token is stored in a cookie for subsequent requests. The token is required for all protected endpoints. The token is used as an identifier for sessions in the database. On logout, the session will be deleted from the database, invalidating the token before its natural expiration. Tokens are rotated after token halflife. Account will be locked after too many failed attempts on the /login endpoint.
