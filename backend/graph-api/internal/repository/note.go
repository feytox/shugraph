package repository

import (
	"context"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository/db"
)

type PgNoteRepository struct {
	queries *db.Queries
}

func NewPgNoteRepository(pool db.DBTX) *PgNoteRepository {
	return &PgNoteRepository{
		queries: db.New(pool),
	}
}

func (p *PgNoteRepository) Create(ctx context.Context, note domain.Note) (domain.Note, error) {
	record := db.CreateNoteParams{
		ID:   note.ID,
		Text: note.Text,
	}
	noteDb, err := p.queries.CreateNote(ctx, record)
	if err != nil {
		return domain.Note{}, err
	}
	return domain.Note{
		ID:        noteDb.ID,
		Text:      noteDb.Text,
		CreatedAt: noteDb.CreatedAt,
	}, err
}

func (p *PgNoteRepository) List(ctx context.Context) ([]domain.Note, error) {
	noteRecords, err := p.queries.GetNotes(ctx)
	if err != nil {
		return nil, err
	}
	notes := make([]domain.Note, 0, len(noteRecords))
	for _, record := range noteRecords {
		note := domain.Note{
			ID:        record.ID,
			Text:      record.Text,
			CreatedAt: record.CreatedAt,
		}
		notes = append(notes, note)
	}
	return notes, nil
}
