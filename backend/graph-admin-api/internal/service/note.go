package service

import (
	"context"
	"fmt"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

type graphAPI interface {
	CreateNote(ctx context.Context, text string) (string, error)
}

type noteRepository interface {
	Create(ctx context.Context, graphNoteID, text string) (domain.ProxiedNote, error)
}

type NoteService struct {
	graph graphAPI
	notes noteRepository
}

func NewNoteService(graph graphAPI, notes noteRepository) *NoteService {
	return &NoteService{graph: graph, notes: notes}
}

func (s *NoteService) Proxy(ctx context.Context, text string) (domain.ProxiedNote, error) {
	graphNoteID, err := s.graph.CreateNote(ctx, text)
	if err != nil {
		return domain.ProxiedNote{}, fmt.Errorf("create note in graph-api: %w", err)
	}

	note, err := s.notes.Create(ctx, graphNoteID, text)
	if err != nil {
		return domain.ProxiedNote{}, fmt.Errorf("save proxied note: %w", err)
	}

	return note, nil
}
