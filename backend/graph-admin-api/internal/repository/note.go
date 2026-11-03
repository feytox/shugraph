package repository

import (
	"context"
	"fmt"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/repository/db"
)

type PgNoteRepository struct {
	queries *db.Queries
}

func NewPgNoteRepository(conn db.DBTX) *PgNoteRepository {
	return &PgNoteRepository{queries: db.New(conn)}
}

func (r *PgNoteRepository) Create(ctx context.Context, graphNoteID, text string) (domain.ProxiedNote, error) {
	row, err := r.queries.CreateProxiedNote(ctx, db.CreateProxiedNoteParams{
		GraphNoteID: graphNoteID,
		Text:        text,
	})
	if err != nil {
		return domain.ProxiedNote{}, fmt.Errorf("insert proxied note: %w", err)
	}

	return domain.ProxiedNote{
		ID:          row.ID,
		GraphNoteID: row.GraphNoteID,
		Text:        row.Text,
		CreatedAt:   row.CreatedAt,
	}, nil
}
