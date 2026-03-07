package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shaftoe/savetoink/backend/internal/logging"
	"github.com/shaftoe/savetoink/backend/internal/model"
)

// writeJSONError writes an error response with the given status code and error message.
func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: err.Error()})
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
	writeJSONError(w, http.StatusInternalServerError, err)
}
