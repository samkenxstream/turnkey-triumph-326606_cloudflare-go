package cloudflare

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Error messages.
const (
	errEmptyCredentials          = "invalid credentials: key & email must not be empty" //nolint:gosec,unused
	errEmptyAPIToken             = "invalid credentials: API Token must not be empty"   //nolint:gosec,unused
	errInternalServiceError      = "internal service error"
	errMakeRequestError          = "error from makeRequest"
	errUnmarshalError            = "error unmarshalling the JSON response"
	errUnmarshalErrorBody        = "error unmarshalling the JSON response error body"
	errRequestNotSuccessful      = "error reported by API"
	errMissingAccountID          = "account ID is empty and must be provided"
	errOperationStillRunning     = "bulk operation did not finish before timeout"
	errOperationUnexpectedStatus = "bulk operation returned an unexpected status"
	errResultInfo                = "incorrect pagination info (result_info) in responses"
	errManualPagination          = "unexpected pagination options passed to functions that handle pagination automatically"
	errInvalidZoneIdentifer      = "invalid zone identifier: %s"
)

var (
	ErrMissingAccountID          = errors.New("required missing account ID")
	ErrMissingZoneID             = errors.New("required missing zone ID")
	ErrMissingResourceIdentifier = errors.New("required missing resource identifier")
)

type ErrorType string

const (
	ErrorTypeRequest        ErrorType = "request"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeService        ErrorType = "service"
)

type Error struct {
	// The classification of error encountered.
	Type ErrorType

	// StatusCode is the HTTP status code from the response.
	StatusCode int

	// Errors is all of the error messages and codes, combined.
	Errors []ResponseInfo

	// ErrorCodes is a list of all the error codes.
	ErrorCodes []int

	// ErrorMessages is a list of all the error codes.
	ErrorMessages []string

	// RayID is the internal identifier for the request that was made.
	RayID string
}

func (e Error) Error() string {
	var errString string
	errMessages := []string{}
	for _, err := range e.Errors {
		m := ""
		if err.Message != "" {
			m += err.Message
		}

		if err.Code != 0 {
			m += fmt.Sprintf(" (%d)", err.Code)
		}

		errMessages = append(errMessages, m)
	}

	return errString + strings.Join(errMessages, ", ")
}

// RequestError is for 4xx errors that we encounter not covered elsewhere
// (generally bad payloads).
type RequestError struct {
	cloudflareError *Error
}

func (e RequestError) Error() string {
	return e.cloudflareError.Error()
}

func (e RequestError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e RequestError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e RequestError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e RequestError) RayID() string {
	return e.cloudflareError.RayID
}

func (e RequestError) Type() ErrorType {
	return e.cloudflareError.Type
}

// RatelimitError is for HTTP 429s where the service is telling the client to
// slow down.
type RatelimitError struct {
	cloudflareError *Error
}

func (e RatelimitError) Error() string {
	return e.cloudflareError.Error()
}

func (e RatelimitError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e RatelimitError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e RatelimitError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e RatelimitError) RayID() string {
	return e.cloudflareError.RayID
}

func (e RatelimitError) Type() ErrorType {
	return e.cloudflareError.Type
}

// ServiceError is a handler for 5xx errors returned to the client.
type ServiceError struct {
	cloudflareError *Error
}

func (e ServiceError) Error() string {
	return e.cloudflareError.Error()
}

func (e ServiceError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e ServiceError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e ServiceError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e ServiceError) RayID() string {
	return e.cloudflareError.RayID
}

func (e ServiceError) Type() ErrorType {
	return e.cloudflareError.Type
}

// AuthenticationError is for HTTP 401 responses.
type AuthenticationError struct {
	cloudflareError *Error
}

func (e AuthenticationError) Error() string {
	return e.cloudflareError.Error()
}

func (e AuthenticationError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e AuthenticationError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e AuthenticationError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e AuthenticationError) RayID() string {
	return e.cloudflareError.RayID
}

func (e AuthenticationError) Type() ErrorType {
	return e.cloudflareError.Type
}

// AuthorizationError is for HTTP 403 responses.
type AuthorizationError struct {
	cloudflareError *Error
}

func (e AuthorizationError) Error() string {
	return e.cloudflareError.Error()
}

func (e AuthorizationError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e AuthorizationError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e AuthorizationError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e AuthorizationError) RayID() string {
	return e.cloudflareError.RayID
}

func (e AuthorizationError) Type() ErrorType {
	return e.cloudflareError.Type
}

// NotFoundError is for HTTP 404 responses.
type NotFoundError struct {
	cloudflareError *Error
}

func (e NotFoundError) Error() string {
	return e.cloudflareError.Error()
}

func (e NotFoundError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e NotFoundError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e NotFoundError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e NotFoundError) RayID() string {
	return e.cloudflareError.RayID
}

func (e NotFoundError) Type() ErrorType {
	return e.cloudflareError.Type
}

// ClientError returns a boolean whether or not the raised error was caused by
// something client side.
func (e *Error) ClientError() bool {
	return e.StatusCode >= http.StatusBadRequest &&
		e.StatusCode < http.StatusInternalServerError
}

// ClientRateLimited returns a boolean whether or not the raised error was
// caused by too many requests from the client.
func (e *Error) ClientRateLimited() bool {
	return e.Type == ErrorTypeRateLimit
}

// InternalErrorCodeIs returns a boolean whether or not the desired internal
// error code is present in `e.InternalErrorCodes`.
func (e *Error) InternalErrorCodeIs(code int) bool {
	for _, errCode := range e.ErrorCodes {
		if errCode == code {
			return true
		}
	}

	return false
}

// ErrorMessageContains returns a boolean whether or not a substring exists in
// any of the `e.ErrorMessages` slice entries.
func (e *Error) ErrorMessageContains(s string) bool {
	for _, errMsg := range e.ErrorMessages {
		if strings.Contains(errMsg, s) {
			return true
		}
	}
	return false
}
