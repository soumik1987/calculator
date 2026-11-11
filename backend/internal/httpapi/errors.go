package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/soumik1987/sami-calculator/backend/internal/calculator"
)

// ErrorCode is a stable, machine-readable error identifier returned to clients.
type ErrorCode string

// Error codes returned in the "code" field of error responses.
const (
	CodeInvalidRequest       ErrorCode = "invalid_request"
	CodeInvalidExpression    ErrorCode = "invalid_expression"
	CodeRequestTooLarge      ErrorCode = "request_too_large"
	CodeUnsupportedMediaType ErrorCode = "unsupported_media_type"
	CodeDivisionByZero       ErrorCode = "division_by_zero"
	CodeNegativeSquareRoot   ErrorCode = "negative_square_root"
	CodeNonPositiveLogarithm ErrorCode = "non_positive_logarithm"
	CodeOutOfRange           ErrorCode = "out_of_range"
	CodeUndefinedResult      ErrorCode = "undefined_result"
	CodeInternal             ErrorCode = "internal_error"
)

const (
	msgEmptyBody            = "Request body is empty."
	msgInvalidJSON          = "Request body is not valid JSON."
	msgNotAnObject          = "Request body must be a JSON object."
	msgMultipleJSONValues   = "Request body must contain a single JSON object."
	msgUnknownField         = "Unknown field %s."
	msgWrongType            = "Field %q must be a string."
	msgTooLarge             = "Request body must be at most %d bytes."
	msgUnsupportedMediaType = "Content-Type must be application/json."
	msgInternal             = "Something went wrong. Please try again."
)

type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type errorMapping struct {
	err     error
	status  int
	code    ErrorCode
	message string
}

var domainErrors = []errorMapping{
	{calculator.ErrDivisionByZero, http.StatusUnprocessableEntity, CodeDivisionByZero, "Cannot divide by zero."},
	{calculator.ErrNegativeSquareRoot, http.StatusUnprocessableEntity, CodeNegativeSquareRoot, "Square root needs a number that is zero or greater."},
	{calculator.ErrNonPositiveLogarithm, http.StatusUnprocessableEntity, CodeNonPositiveLogarithm, "Logarithm needs a number greater than zero."},
	{calculator.ErrOutOfRange, http.StatusUnprocessableEntity, CodeOutOfRange, "The result is outside the supported number range."},
	{calculator.ErrUndefinedResult, http.StatusUnprocessableEntity, CodeUndefinedResult, "The result is not a real number."},
}

func (h *handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var syntaxErr *calculator.SyntaxError
	if errors.As(err, &syntaxErr) {
		h.fail(w, r, http.StatusBadRequest, CodeInvalidExpression, syntaxErr.Error())
		return
	}
	for _, d := range domainErrors {
		if errors.Is(err, d.err) {
			h.fail(w, r, d.status, d.code, d.message)
			return
		}
	}
	h.logger.Error("request failed", "error", err, "traceId", traceID(r))
	h.fail(w, r, http.StatusInternalServerError, CodeInternal, msgInternal)
}

// fail writes an error response and tags the request's span with the error code.
func (h *handler) fail(w http.ResponseWriter, r *http.Request, status int, code ErrorCode, message string) {
	recordErrorType(r, code)
	writeError(w, status, code, message)
}

func writeError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set(headerContentType, mediaTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
