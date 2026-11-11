// Package httpapi exposes the calculator service over HTTP/JSON.
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
)

const maxBodyBytes = 4 << 10

// Evaluator computes an expression. In production it is calculator.Evaluate; tests pass fakes.
type Evaluator func(expression string) (float64, error)

type handler struct {
	evaluate Evaluator
	logger   *slog.Logger
}

// NewHandler returns the routes of the calculator API.
func NewHandler(evaluate Evaluator, logger *slog.Logger) http.Handler {
	h := &handler{evaluate: evaluate, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc(http.MethodGet+" "+pathHealth, h.health)
	mux.HandleFunc(http.MethodPost+" "+pathCalculations, h.calculate)
	return mux
}

type healthStatus string

const healthStatusOK healthStatus = "ok"

type healthResponse struct {
	Status healthStatus `json:"status"`
}

type calculateRequest struct {
	Expression string `json:"expression"`
}

type calculationResponse struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: healthStatusOK})
}

func (h *handler) calculate(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		h.fail(w, r, http.StatusUnsupportedMediaType, CodeUnsupportedMediaType, msgUnsupportedMediaType)
		return
	}

	var req calculateRequest
	if reqErr := decodeJSON(w, r, &req); reqErr != nil {
		h.fail(w, r, reqErr.status, reqErr.code, reqErr.message)
		return
	}

	result, err := h.evaluate(req.Expression)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, calculationResponse{Expression: req.Expression, Result: result})
}

func isJSON(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(headerContentType))
	return err == nil && mediaType == mediaTypeJSON
}

type requestError struct {
	status  int
	code    ErrorCode
	message string
}

func badRequest(format string, args ...any) *requestError {
	return &requestError{status: http.StatusBadRequest, code: CodeInvalidRequest, message: fmt.Sprintf(format, args...)}
}

const unknownFieldPrefix = "json: unknown field "

// decodeJSON reads exactly one JSON object into dst and explains what is wrong if it can't.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) *requestError {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err == nil {
		if !errors.Is(dec.Decode(&struct{}{}), io.EOF) {
			return badRequest(msgMultipleJSONValues)
		}
		return nil
	}

	var tooLarge *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &tooLarge):
		return &requestError{
			status:  http.StatusRequestEntityTooLarge,
			code:    CodeRequestTooLarge,
			message: fmt.Sprintf(msgTooLarge, maxBodyBytes),
		}
	case errors.Is(err, io.EOF):
		return badRequest(msgEmptyBody)
	case errors.As(err, &typeErr) && typeErr.Field == "":
		return badRequest(msgNotAnObject)
	case errors.As(err, &typeErr):
		return badRequest(msgWrongType, typeErr.Field)
	case strings.HasPrefix(err.Error(), unknownFieldPrefix):
		return badRequest(msgUnknownField, strings.TrimPrefix(err.Error(), unknownFieldPrefix))
	default:
		return badRequest(msgInvalidJSON)
	}
}
