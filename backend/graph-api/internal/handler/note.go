package handler

import (
	"context"
	"graph-api/internal/service"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type NoteHandler struct {
	service *service.NoteService
}

func NewNoteHandler(service *service.NoteService) *NoteHandler {
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
	Id        string    `json:"id"`
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
		return nil, huma.Error503ServiceUnavailable(err.Error(), err)
	}
	response := CreateNoteOutput{}
	response.Body = Note{
		Id:        note.ID.String(),
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
	notes, err := h.service.GetAllNote(ctx)
	if err != nil {
		return nil, huma.Error503ServiceUnavailable(err.Error(), err)
	}
	response := AllNoteOutput{}
	response.Body.Notes = make([]Note, 0, len(notes))
	for _, note := range notes {
		response.Body.Notes = append(response.Body.Notes, Note{
			Id:        note.ID.String(),
			Text:      note.Text,
			CreatedAt: note.CreatedAt.UTC(),
		})
	}
	return &response, nil
}
