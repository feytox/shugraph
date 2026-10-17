//go:build integration

package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/config"
)

func TestProxyNoteIntegration(t *testing.T) {
	graphAPIStatus := &atomic.Int32{}
	srv := startServer(t, startFakeGraphAPI(t, graphAPIStatus))

	t.Run("readyz", func(t *testing.T) {
		rec := request(t, srv, http.MethodGet, "/readyz", "")

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("note is saved", func(t *testing.T) {
		graphAPIStatus.Store(http.StatusCreated)

		rec := request(t, srv, http.MethodPost, "/api/v1/notes/proxy", `{"text":"hello"}`)
		require.Equal(t, http.StatusCreated, rec.Code)

		var resp struct {
			ID int64 `json:"id"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

		var graphNoteID, text string

		err := srv.pool.QueryRow(t.Context(), "SELECT graph_note_id, text FROM proxied_notes WHERE id = $1", resp.ID).Scan(&graphNoteID, &text)

		require.NoError(t, err)
		assert.Equal(t, "note-1", graphNoteID)
		assert.Equal(t, "hello", text)
	})

	t.Run("graph-api failure is not saved", func(t *testing.T) {
		graphAPIStatus.Store(http.StatusServiceUnavailable)

		before := countNotes(t, srv)

		rec := request(t, srv, http.MethodPost, "/api/v1/notes/proxy", `{"text":"lost"}`)

		assert.Equal(t, http.StatusBadGateway, rec.Code)
		assert.Equal(t, before, countNotes(t, srv))
	})
}

func startFakeGraphAPI(t *testing.T, status *atomic.Int32) string {
	t.Helper()

	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(status.Load()))
		_, _ = io.WriteString(w, `{"id":"note-1","text":"hello","created_at":"2026-10-16T10:00:00Z"}`)
	}))
	t.Cleanup(graph.Close)

	return graph.URL
}

func startServer(t *testing.T, graphAPIURL string) *server {
	t.Helper()

	ctx := t.Context()

	pg, err := postgres.Run(ctx, "postgres:18-alpine", postgres.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, pg)
	require.NoError(t, err)

	databaseURL, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	cfg, err := env.ParseAsWithOptions[config.Serve](env.Options{Environment: map[string]string{
		"DATABASE_URL":  databaseURL,
		"GRAPH_API_URL": graphAPIURL,
	}})
	require.NoError(t, err)

	log := slog.New(slog.DiscardHandler)

	srv, err := newServer(ctx, cfg, log)
	require.NoError(t, err)
	t.Cleanup(srv.close)

	require.NoError(t, migrate(ctx, srv.pool, log))

	return srv
}

func request(t *testing.T, srv *server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	return rec
}

func countNotes(t *testing.T, srv *server) int {
	t.Helper()

	var count int
	require.NoError(t, srv.pool.QueryRow(t.Context(), "SELECT count(*) FROM proxied_notes").Scan(&count))

	return count
}
