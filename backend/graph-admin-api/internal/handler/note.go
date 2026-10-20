package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

const notesTag = "Notes"

type NoteService interface {
	Proxy(ctx context.Context, text string) (domain.ProxiedNote, error)
}

type ProxiedNote struct {
	ID          int64     `json:"id"`
	GraphNoteID string    `json:"graph_note_id" doc:"Note ID in graph-api"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProxyNoteInput struct {
	Body struct {
		Text string `json:"text" minLength:"1"`
	}
}

type ProxyNoteOutput struct {
	Body ProxiedNote
}

type noteHandlers struct {
	notes NoteService
	log   *slog.Logger
}

func registerNotes(api huma.API, h *noteHandlers) {
	huma.Register(api, huma.Operation{
		OperationID:   "proxy-note",
		Summary:       "Create a note in graph-api",
		Method:        http.MethodPost,
		Path:          "/api/v1/notes/proxy",
		Tags:          []string{notesTag},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError, http.StatusBadGateway},
	}, h.proxy)
}

func (h *noteHandlers) proxy(ctx context.Context, in *ProxyNoteInput) (*ProxyNoteOutput, error) {
	note, err := h.notes.Proxy(ctx, in.Body.Text)
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	return &ProxyNoteOutput{Body: ProxiedNote{
		ID:          note.ID,
		GraphNoteID: note.GraphNoteID,
		Text:        note.Text,
		CreatedAt:   note.CreatedAt.UTC(),
	}}, nil
}
