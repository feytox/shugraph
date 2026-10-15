package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"

	"github.com/danielgtaylor/huma/v2"
)

type noteService interface {
	CreateNote(ctx context.Context, text string) (domain.Note, error)
	List(ctx context.Context) ([]domain.Note, error)
}

type NoteHandler struct {
	service noteService
}

func NewNoteHandler(service noteService) *NoteHandler {
	return &NoteHandler{service: service}
}

func (h *NoteHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-note",
		Summary:       "Create a new note",
		Method:        http.MethodPost,
		Path:          "/api/v1/notes",
		DefaultStatus: http.StatusCreated,
	}, h.CreateNote)

	huma.Register(api, huma.Operation{
		OperationID: "get-all-notes",
		Summary:     "Retrieve all notes",
		Method:      http.MethodGet,
		Path:        "/api/v1/notes",
	}, h.GetAllNote)
}

type Note struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateNoteInput struct {
	Body struct {
		Text string `json:"text" minLength:"1" doc:"Текст заметки"`
	}
}

type CreateNoteOutput struct {
	Body Note
}

func (h *NoteHandler) CreateNote(ctx context.Context, request *CreateNoteInput) (*CreateNoteOutput, error) {
	note, err := h.service.CreateNote(ctx, request.Body.Text)
	if err != nil {
		slog.Error("database is unavailable", "err", err)
		return nil, huma.Error503ServiceUnavailable("database is unavailable")
	}
	response := CreateNoteOutput{}
	response.Body = Note{
		ID:        note.ID.String(),
		Text:      note.Text,
		CreatedAt: note.CreatedAt,
	}
	return &response, nil
}

type AllNoteOutput struct {
	Body struct {
		Notes []Note `json:"notes"`
	}
}

func (h *NoteHandler) GetAllNote(ctx context.Context, _ *struct{}) (*AllNoteOutput, error) {
	notes, err := h.service.List(ctx)
	if err != nil {
		slog.Error("database is unavailable", "err", err)
		return nil, huma.Error503ServiceUnavailable("database is unavailable")
	}
	response := AllNoteOutput{}
	response.Body.Notes = make([]Note, 0, len(notes))
	for _, note := range notes {
		response.Body.Notes = append(response.Body.Notes, Note{
			ID:        note.ID.String(),
			Text:      note.Text,
			CreatedAt: note.CreatedAt.UTC(),
		})
	}
	return &response, nil
}
