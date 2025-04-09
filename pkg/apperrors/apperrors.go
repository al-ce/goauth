package apperrors

import (
	"errors"
)

var New = errors.New

var (
	ErrAccountIsLocked               = New("Account is locked")
	ErrCouldNotIncrementFailedLogins = New("Could not increment users.failed_login_attempts")
	ErrCouldNotUpdateUser            = New("Tried to update user but no changes were made")
	ErrDuplicateEmail                = New("Duplicate email")
	ErrEmailIsEmpty                  = New("Email is empty")
	ErrEmailMaxLength                = New("Email exceeds max length of 254 characters")
	ErrFailedUserUpade               = New("Could not update user")
	ErrInvalidLogin                  = New("Invalid login credentials")
	ErrPasswordIsEmpty               = New("Password is empty")
	ErrSessionAlreadyExists          = New("Session already exists")
	ErrSessionIsNil                  = New("Session is nil")
	ErrTokenGeneration               = New("Could not generate token")
	ErrTokenIsEmpty                  = New("Token is empty")
	ErrUserIdEmpty                   = New("User ID is empty")
	ErrUserIsNil                     = New("User is nil")
	ErrUserNotFound                  = New("User not found")
)
