package web

import (
	"encoding/json"
	"net/http"
)

// HealthResponse: returned by the health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse: returned when an error occurs
type ErrorResponse struct {
	Message string `json:"message"`
}

// JSON writes the given payload as a JSON response with the specified HTTP status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Error writes an error message as a JSON response with the specified HTTP status code.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Message: message})
}
