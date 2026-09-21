package crayfi

import "fmt"

// CrayException is the base interface for all Cray exceptions
type CrayException interface {
	error
	IsCrayException() bool
}

type baseException struct {
	Message string
}

func (e *baseException) Error() string {
	return e.Message
}

func (e *baseException) IsCrayException() bool {
	return true
}

// AuthenticationException is thrown when API key is missing or invalid
type AuthenticationException struct {
	baseException
}

func NewAuthenticationException(message string) *AuthenticationException {
	return &AuthenticationException{baseException{Message: message}}
}

type SignatureException struct {
	AuthenticationException
	Code string
}

func NewSignatureException(message, code string) *SignatureException {
	return &SignatureException{AuthenticationException: *NewAuthenticationException(message), Code: code}
}

// ValidationException is thrown when input validation fails
type ValidationException struct {
	baseException
}

func NewValidationException(message string) *ValidationException {
	return &ValidationException{baseException{Message: message}}
}

// TimeoutException is thrown when a request times out
type TimeoutException struct {
	baseException
}

func NewTimeoutException(message string) *TimeoutException {
	return &TimeoutException{baseException{Message: message}}
}

// APIException is thrown when the API returns an error response
type APIException struct {
	baseException
	StatusCode int
	Body       interface{}
}

func NewAPIException(message string, statusCode int, body interface{}) *APIException {
	return &APIException{
		baseException: baseException{Message: message},
		StatusCode:    statusCode,
		Body:          body,
	}
}

func (e *APIException) Error() string {
	return fmt.Sprintf("Cray API Error (%d): %s", e.StatusCode, e.Message)
}
