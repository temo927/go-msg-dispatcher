package http

import (
	"net/http"
	"strconv"

	"github.com/temo927/go-msg-dispatcher/internal/app"
	"github.com/temo927/go-msg-dispatcher/internal/domain"
)

type Handlers struct {
	Scheduler *app.Scheduler
	Repo      domain.MessagesRepo
}

func NewHandlers(scheduler *app.Scheduler, repo domain.MessagesRepo) *Handlers {
	return &Handlers{Scheduler: scheduler, Repo: repo}
}

func (h *Handlers) StartScheduler(w http.ResponseWriter, r *http.Request) {
	if err := h.Scheduler.Start(r.Context()); err != nil {
		switch err {
		case app.ErrAlreadyRunning:
			JSONError(w, http.StatusConflict, err.Error())
		default:
			JSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	JSONSuccess(w, http.StatusOK, map[string]string{"message": "scheduler started"})
}

func (h *Handlers) StopScheduler(w http.ResponseWriter, r *http.Request) {
	if err := h.Scheduler.Stop(); err != nil {
		switch err {
		case app.ErrNotRunning:
			JSONError(w, http.StatusConflict, err.Error())
		default:
			JSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	JSONSuccess(w, http.StatusOK, map[string]string{"message": "scheduler stopped"})
}

func (h *Handlers) ListSent(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	q := r.URL.Query()
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	msgs, err := h.Repo.ListSent(r.Context(), limit, offset)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		resp = append(resp, map[string]any{
			"id":                 m.ID,
			"to_phone":           m.ToPhone,
			"content":            m.Content,
			"provider_message_id": m.ProviderMessageID,
			"sent_at":            m.SentAt,
		})
	}
	JSONSuccess(w, http.StatusOK, map[string]any{"items": resp, "count": len(resp)})
}
