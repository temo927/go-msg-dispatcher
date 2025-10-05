package http

import (
	"net/http"
)

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/scheduler/start", h.StartScheduler)
	mux.HandleFunc("/api/v1/scheduler/stop", h.StopScheduler)
	mux.HandleFunc("/api/v1/messages/sent", h.ListSent)

	RegisterSwagger(mux, "internal/transport/http/swagger")

	return RecoverMiddleware(RequestLogger(mux))
}
