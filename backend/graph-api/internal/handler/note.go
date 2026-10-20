package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

const notesTag = "Notes"

type NoteService interface {
	CreateNote(ctx context.Context, text string) (domain.Note, error)
	List(ctx context.Context) ([]domain.Note, error)
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

type AllNoteOutput struct {
	Body struct {
		Notes []Note `json:"notes"`
	}
}

type noteHandlers struct {
	notes NoteService
	log   *slog.Logger
}

func registerNotes(api huma.API, h *noteHandlers) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-note",
		Summary:       "Create a new note",
		Method:        http.MethodPost,
		Path:          "/api/v1/notes",
		Tags:          []string{notesTag},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError},
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "get-all-notes",
		Summary:     "Retrieve all notes",
		Method:      http.MethodGet,
		Path:        "/api/v1/notes",
		Tags:        []string{notesTag},
		Errors:      []int{http.StatusInternalServerError},
	}, h.list)
}

func (h *noteHandlers) create(ctx context.Context, in *CreateNoteInput) (*CreateNoteOutput, error) {
	note, err := h.notes.CreateNote(ctx, in.Body.Text)
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	return &CreateNoteOutput{Body: Note{
		ID:        note.ID.String(),
		Text:      note.Text,
		CreatedAt: note.CreatedAt,
	}}, nil
}

func (h *noteHandlers) list(ctx context.Context, _ *struct{}) (*AllNoteOutput, error) {
	notes, err := h.notes.List(ctx)
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	out := &AllNoteOutput{}
	out.Body.Notes = make([]Note, 0, len(notes))

	for _, note := range notes {
		out.Body.Notes = append(out.Body.Notes, Note{
			ID:        note.ID.String(),
			Text:      note.Text,
			CreatedAt: note.CreatedAt.UTC(),
		})
	}

	return out, nil
}
