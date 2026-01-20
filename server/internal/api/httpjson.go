package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Request string `json:"requestId,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	requestID, _ := r.Context().Value(ctxKeyRequestID{}).(string)
	writeJSON(w, status, ErrorResponse{Error: msg, Request: requestID})
}
