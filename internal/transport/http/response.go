package http

import (
	"encoding/json"
	"net/http"

	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
)

type responseEnvelope struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func JSONSuccess(w http.ResponseWriter, code int, data interface{}) {
	resp := responseEnvelope{
		Status: "ok",
		Data:   data,
	}
	writeJSON(w, code, resp)
}

func JSONError(w http.ResponseWriter, code int, message string) {
	resp := responseEnvelope{
		Status: "error",
		Error:  message,
	}
	writeJSON(w, code, resp)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Logger.Error("failed to encode JSON response", "err", err)
		http.Error(w, `{"status":"error","error":"internal error"}`, http.StatusInternalServerError)
	}
}
