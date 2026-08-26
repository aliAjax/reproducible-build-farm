package api

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Error{Code: code, Message: message, RequestID: requestID})
}
