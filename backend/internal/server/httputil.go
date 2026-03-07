package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/shaftoe/savetoink/backend/internal/apperrors"
	"github.com/shaftoe/savetoink/backend/internal/logging"
	"github.com/shaftoe/savetoink/backend/internal/model"
)

// writeJSONError writes an error response with the given status code and error message.
func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
}

// statusCodeForError returns the appropriate HTTP status code for the given error.
func statusCodeForError(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperrors.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// decodeAndValidateRequest decodes JSON from request body and handles errors.
func decodeAndValidateRequest(w http.ResponseWriter, r *http.Request, req any) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		decodeErr := fmt.Errorf("failed to decode request body: %w", err)
		writeJSONError(w, http.StatusBadRequest, decodeErr)
		return decodeErr
	}
	return nil
}

// handleServiceError logs error and writes appropriate response.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error, context string) {
	logging.AddRequestError(r.Context(), fmt.Errorf("%s: %w", context, err))
	writeJSONError(w, statusCodeForError(err), err)
}
