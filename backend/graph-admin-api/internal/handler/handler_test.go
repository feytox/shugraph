package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

type noteServiceFunc func(ctx context.Context, text string) (domain.ProxiedNote, error)

func (f noteServiceFunc) Proxy(ctx context.Context, text string) (domain.ProxiedNote, error) {
	return f(ctx, text)
}

type pingerFunc func(ctx context.Context) error

func (f pingerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func TestProxyNote(t *testing.T) {
	createdAt := time.Date(2026, 10, 16, 15, 0, 0, 0, time.FixedZone("UTC+5", 5*60*60))
	notes := noteServiceFunc(func(_ context.Context, text string) (domain.ProxiedNote, error) {
		return domain.ProxiedNote{ID: 7, GraphNoteID: "note-1", Text: text, CreatedAt: createdAt}, nil
	})

	rec := serve(t, Deps{Notes: notes}, http.MethodPost, "/api/v1/notes/proxy", `{"text":"hello"}`)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"id":7,"graph_note_id":"note-1","text":"hello","created_at":"2026-10-16T10:00:00Z"}`, rec.Body.String())
}

func TestProxyNoteErrors(t *testing.T) {
	tests := map[string]struct {
		err        error
		wantStatus int
		wantDetail string
	}{
		"graph-api unavailable": {
			err:        fmt.Errorf("create note in graph-api: %w", domain.ErrGraphAPIUnavailable),
			wantStatus: http.StatusBadGateway,
			wantDetail: "graph-api is unavailable",
		},
		"unknown error is hidden": {
			err:        errors.New("save proxied note: connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantDetail: "internal error",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			notes := noteServiceFunc(func(context.Context, string) (domain.ProxiedNote, error) {
				return domain.ProxiedNote{}, tt.err
			})

			rec := serve(t, Deps{Notes: notes}, http.MethodPost, "/api/v1/notes/proxy", `{"text":"hello"}`)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), `"detail":"`+tt.wantDetail+`"`)
		})
	}
}

func TestProxyNoteEmptyText(t *testing.T) {
	notes := noteServiceFunc(func(context.Context, string) (domain.ProxiedNote, error) {
		t.Error("service must not be called")
		return domain.ProxiedNote{}, nil
	})

	rec := serve(t, Deps{Notes: notes}, http.MethodPost, "/api/v1/notes/proxy", `{"text":""}`)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestReadyz(t *testing.T) {
	tests := map[string]struct {
		pingErr    error
		wantStatus int
	}{
		"database is up":   {wantStatus: http.StatusOK},
		"database is down": {pingErr: errors.New("connection refused"), wantStatus: http.StatusServiceUnavailable},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			database := pingerFunc(func(context.Context) error { return tt.pingErr })

			rec := serve(t, Deps{Database: database}, http.MethodGet, "/readyz", "")

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func serve(t *testing.T, deps Deps, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	deps.Log = slog.New(slog.DiscardHandler)
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	New(deps).ServeHTTP(rec, req)

	return rec
}
