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
)
