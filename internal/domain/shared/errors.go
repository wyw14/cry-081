package shared

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrForbidden       = errors.New("operation forbidden")
	ErrConflict        = errors.New("resource conflict")
	ErrInvalidState    = errors.New("invalid state transition")
	ErrVersionConflict = errors.New("optimistic version conflict")
	ErrValidation      = errors.New("validation failed")
	ErrUnauthenticated = errors.New("authentication required")
)

type FieldViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type DomainError struct {
	Code       string
	Message    string
	Cause      error
	Violations []FieldViolation
}

func (e *DomainError) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *DomainError) Unwrap() error { return e.Cause }

func NewError(code, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

func ValidationError(violations ...FieldViolation) *DomainError {
	return &DomainError{
		Code:       "VALIDATION_FAILED",
		Message:    "request validation failed",
		Cause:      ErrValidation,
		Violations: violations,
	}
}
