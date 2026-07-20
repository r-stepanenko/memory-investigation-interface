package http

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(
		ErrorResponse{
			Error: ErrorDetails{
				Code:    code,
				Message: message,
			},
		},
	)
}