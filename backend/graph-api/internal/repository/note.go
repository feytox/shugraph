package repository

import (
	"context"
	"graph-api/internal/domain"
	"graph-api/internal/repository/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository interface {
	Create(ctx context.Context, product *domain.Note) error
	GetAll(ctx context.Context) ([]domain.Note, error)
}

type PgNoteRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewPgNoteRepository(pool *pgxpool.Pool) *PgNoteRepository {
	return &PgNoteRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (p PgNoteRepository) Create(ctx context.Context, note *domain.Note) error {
	record := db.CreateNoteParams{
		ID:        note.Id,
		Text:      note.Text,
		CreatedAt: note.CreatedAt,
	}
	err := p.queries.CreateNote(ctx, record)
	return err
}

func (p PgNoteRepository) GetAll(ctx context.Context) ([]domain.Note, error) {
	noteRecords, err := p.queries.GetNotes(ctx)
	if err != nil {
		return nil, err
	}
	notes := make([]domain.Note, 0, len(noteRecords))
	for _, record := range noteRecords {
		note := domain.Note{
			Id:        record.ID,
			Text:      record.Text,
			CreatedAt: record.CreatedAt,
		}
		notes = append(notes, note)
	}
	return notes, nil
}
