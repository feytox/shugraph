package service

import (
	"context"
	"graph-api/internal/domain"
	"graph-api/internal/repository"
	"time"

	"github.com/google/uuid"
)

type NoteService struct {
	noteRepository repository.NoteRepository
}

func NewNoteService(noteRepository repository.NoteRepository) *NoteService {
	return &NoteService{noteRepository: noteRepository}
}

func (s *NoteService) CreateNote(ctx context.Context, text string) (*domain.Note, error) {
	note := &domain.Note{
		Id:        uuid.New(),
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	err := s.noteRepository.Create(ctx, note)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteService) GetAllNote(ctx context.Context) ([]domain.Note, error) {
	notes, err := s.noteRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, note := range notes {
		note.CreatedAt = note.CreatedAt.UTC()
	}
	return notes, nil
}
