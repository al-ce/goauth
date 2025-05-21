# Authentication Service

The `auth` directory contains the code to build the user registration and authentication system for the platform.

## Documentation

See [api.md] for an overview of the available endpoints and the expected request/response format.
Or see the [Swagger](https://swagger.io/docs/) documentation for the endpoints at e.g. `http://{URL}:{PORT}/swagger/index.html`

## Build

Build like a typical Go module:

```bash
go build -o auth ./main.go
```

## Directory Structure

```plaintext
 auth
├──  docs
├──  internal
│   ├──  database
│   ├──  handlers
│   ├──  middleware
│   ├──  models
│   ├──  repository
│   ├──  server
│   ├──  services
│   └──  testutils
├──  pkg
│   └──  apperrors
│   └──  config
│   └──  logger
└──  scripts
```

- `docs`: Contains documentation files related to the authentication system
- `internal`: internal packages that are not meant to be used outside of the `auth` module
    - `database`: code related to database interactions for the authentication system
    - `handlers`: handler functions for HTTP routes
    - `middleware`: middleware used for user/admin authentication
    - `models`: models for database tables `users` and `sessions`, automigrated
    - `repository`: code to perform CRUD and other operations on `users` and `sessions` tables
    - `server`: code to setup and run API server
    - `services`: functions for mediating logic between HTTP handler functions and repository functions
    - `testutils`: utility functions and types for testing the authentication system
- `pkg`: packages that are meant to be used by other modules
    - `apperrors`: custom errors for testing and logging
    - `config`: constants for configuring auth operations
    - `logger`: configuration and setup for logging
- `scripts`: utility scripts for local development and testing of the authentication system

## Dependencies

The `auth` module expects the following environment variables to be set:

- `DATABASE_URL`: The URL of the database to connect to
- `AUTH_SERVER_PORT`: The port to run the http server on
- `DISCUSSION_APP_SESSION_KEY`: The secret key to encrypt the session id
- `CORS_ALLOWED_ORIGINS`: comma separated string of allowed origins e.g. "http://localhost:5173,http://localhost:4173"


See `example.env` or the `watch` command in `justfile` for sample environment variables.

Third party packages are defined in `go.mod` and `go.sum`.
