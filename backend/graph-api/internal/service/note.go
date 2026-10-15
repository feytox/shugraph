package service

import (
	"context"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"

	"github.com/google/uuid"
)

type noteRepository interface {
	Create(ctx context.Context, note domain.Note) (domain.Note, error)
	List(ctx context.Context) ([]domain.Note, error)
}

type NoteService struct {
	noteRepository noteRepository
}

func NewNoteService(noteRepository noteRepository) *NoteService {
	return &NoteService{noteRepository: noteRepository}
}

func (s *NoteService) CreateNote(ctx context.Context, text string) (domain.Note, error) {
	note := domain.Note{
		ID:   uuid.New(),
		Text: text,
	}

	noteDb, err := s.noteRepository.Create(ctx, note)
	if err != nil {
		return domain.Note{}, err
	}
	return noteDb, nil
}

func (s *NoteService) List(ctx context.Context) ([]domain.Note, error) {
	notes, err := s.noteRepository.List(ctx)
	if err != nil {
		return nil, err
	}
	return notes, nil
}
