package graphapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

const noteJSON = `{"id":"note-1","text":"hello","created_at":"2026-10-16T10:00:00Z"}`

func TestCreateNote(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		assert.Equal(t, "POST /api/v1/notes", r.Method+" "+r.URL.Path)
		assert.Equal(t, "req-1", r.Header.Get(middleware.RequestIDHeader))
		assert.JSONEq(t, `{"text":"hello"}`, string(body))

		writeJSON(w, http.StatusCreated, noteJSON)
	}))
	t.Cleanup(graph.Close)

	ctx := context.WithValue(t.Context(), middleware.RequestIDKey, "req-1")
	id, err := newTestClient(t, graph.URL).CreateNote(ctx, "hello")

	require.NoError(t, err)
	assert.Equal(t, "note-1", id)
}

func TestCreateNoteGraphAPIUnavailable(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"error status":       func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
		"200 instead of 201": func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, noteJSON) },
		"timeout":            waitForClientGone,
	}
	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			graph := httptest.NewServer(handler)
			t.Cleanup(graph.Close)

			_, err := newTestClient(t, graph.URL).CreateNote(t.Context(), "hello")

			assert.ErrorIs(t, err, domain.ErrGraphAPIUnavailable)
		})
	}
}

func waitForClientGone(_ http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	<-r.Context().Done()
}

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	client, err := New(baseURL, &http.Client{Timeout: 100 * time.Millisecond})
	require.NoError(t, err)

	return client
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
