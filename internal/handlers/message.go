package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/vkuksa/twatter/internal"
	"go.uber.org/zap"
)

type MessageService interface {
	AddMessage(ctx context.Context, content string)

	GenerateMessageFeed(ctx context.Context) (chan internal.Message, error)
}

type MessageHandler struct {
	logger *zap.Logger
	svc    MessageService
}

func NewMessageHandler(l *zap.Logger, svc MessageService) *MessageHandler {
	return &MessageHandler{
		logger: l,
		svc:    svc,
	}
}

func (m *MessageHandler) Register(r *chi.Mux) {
	r.Post("/add", m.handleAdd)
	r.Get("/feed", m.handleFeed)
}

func (m *MessageHandler) handleAdd(w http.ResponseWriter, r *http.Request) {
	content := r.PostFormValue("content")
	m.svc.AddMessage(r.Context(), content)
	w.WriteHeader(http.StatusAccepted)
}

func (m *MessageHandler) handleFeed(w http.ResponseWriter, r *http.Request) {
	// Set appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	msgChan, err := m.svc.GenerateMessageFeed(r.Context())
	if err != nil {
		m.logger.Error(err.Error())
		renderErrorResponse(w, r, "feed streaming failed", err)
		return
	}

	for msg := range msgChan {
		_, err = w.Write([]byte(msg.String()))
		if err != nil {
			m.logger.Error(err.Error())
			renderErrorResponse(w, r, "feed streaming failed", err)
			return
		}

		// Flush the response writer to ensure the event is sent immediately
		w.(http.Flusher).Flush()
	}

	render.Status(r, http.StatusOK)
}
