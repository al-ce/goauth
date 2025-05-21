package apperrors

import (
	"errors"
)

var New = errors.New

var (
	// Authentication errors
	ErrAccountIsLocked     = New("Account is locked")
	ErrInvalidLogin        = New("Invalid login credentials")
	ErrInvalidTokenFormat  = New("Invalid token format")
	ErrSessionIDGeneration = New("Could not generate token")

	// User registration errors
	ErrDuplicateEmail     = New("User already registered with this email")
	ErrEmailIsEmpty       = New("Email is empty")
	ErrEmailMaxLength     = New("Email exceeds max length of 254 characters")
	ErrEmailFormat        = New("Email format is invalid")
	ErrPasswordComplexity = New("Please use a more complex password! https://xkcd.com/936")

	// Session errors
	ErrSessionAlreadyExists = New("Session already exists")

	// Nil reference argument errors
	ErrDatabaseIsNil     = New("Database is nil")
	ErrSessionIsNil      = New("Session is nil")
	ErrUserIsNil         = New("User is nil")
	ErrSessionRepoIsNil  = New("Session repo is nil")
	ErrUserRepoIsNil     = New("UserRepo is nil")
	ErrUserServiceIsNil  = New("UserService is nil")
	ErrUserHandlerIsNil  = New("UserHandler is nil")
	ErrRepoProviderIsNil = New("RepoProvider is nil")

	// Empty string argument errors
	ErrExpiresAtIsEmpty = New("Expiration time is empty")
	ErrPasswordIsEmpty  = New("Password is empty")
	ErrSessionIdIsEmpty = New("Token is empty")
	ErrUserIdEmpty      = New("User ID is empty")

	// Database errors
	ErrUserNotFound = New("User not found")

	ErrCouldNotIncrementFailedLogins = New("Could not increment users.failed_login_attempts")
	ErrCouldNotUpdateUser            = New("Tried to update user but no changes were made")
	ErrMissingCredentials            = New("Requires both email and password")
)
