package apperrors

import (
	"errors"
)

var New = errors.New

var (
	ErrUserIsNil       = New("User is nil")
	ErrEmailIsEmpty    = New("Email is empty")
	ErrPasswordIsEmpty = New("Password is empty")
	ErrUserNotFound    = New("User not found")
	ErrDuplicateEmail  = New("Duplicate email")
	ErrEmailMaxLength  = New("Email exceeds max length of 254 characters")
	ErrInvalidLogin    = New("Invalid login credentials")
	ErrTokenGeneration = New("Could not generate token")
	ErrUserIdEmpty     = New("User ID is empty")
	ErrFailedUserUpade = New("Could not update user")
	ErrNoChangesMade   = New("Tried to update user but no changes were made")
	ErrSessionIsNil    = New("Session is nil")
	ErrSessionAlreadyExists   = New("Session already exists")
	ErrTokenIsEmpty    = New("Token is empty")
)
