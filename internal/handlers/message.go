package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/vkuksa/twatter/internal"
)

type MessageService interface {
	Add(ctx context.Context, content string) (internal.Message, error)

	GenerateFeed(ctx context.Context) (chan internal.Message, error)
}

type MessageHandler struct {
	svc MessageService
}

func NewMessageHandler(svc MessageService) *MessageHandler {
	return &MessageHandler{
		svc: svc,
	}
}

func (m *MessageHandler) Register(r *chi.Mux) {
	r.Post("/add", m.handleAdd)
	r.Get("/feed", m.handleFeed)
}

func (m *MessageHandler) handleAdd(w http.ResponseWriter, r *http.Request) {
	content := r.PostFormValue("content")
	msg, err := m.svc.Add(r.Context(), content)
	if err != nil {
		renderErrorResponse(w, r, "add failed", err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.HTML(w, r, msg.String())
}

func (m *MessageHandler) handleFeed(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	msgChan, err := m.svc.GenerateFeed(r.Context())
	if err != nil {
		renderErrorResponse(w, r, "feed streaming failed", err)
		return
	}

	for msg := range msgChan {
		_, err = w.Write([]byte(msg.String()))
		if err != nil {
			renderErrorResponse(w, r, "feed streaming failed", err)
		}

		// Flush the response writer to ensure the event is sent immediately
		w.(http.Flusher).Flush()
	}

	render.Status(r, http.StatusOK)
}
