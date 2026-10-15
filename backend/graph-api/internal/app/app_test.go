//go:build integration

package app

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
	"github.com/feytox/shugraph/backend/graph-api/internal/repository"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var serverURL string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("users"),
		postgres.WithUsername("users"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		slog.Error("Could not start postgres", "err", err)
		return 1
	}

	defer func(pgContainer *postgres.PostgresContainer, ctx context.Context) {
		err := pgContainer.Terminate(ctx)
		if err != nil {
			slog.Error("Could not terminate postgres container", "err", err)
		}
	}(pgContainer, ctx)

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		slog.Error("Could not get connection string to postgres", "err", err)
		return 1
	}
	if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
		slog.Error("could not set DATABASE_URL", "err", err)
		return 1
	}
	pool, err := ConnectToPg(ctx, dbURL)
	if err != nil {
		slog.Error("Could not connect to postgres", "err", err)
		return 1
	}
	defer pool.Close()

	repo := repository.NewPgNoteRepository(pool)
	noteService := service.NewNoteService(repo)

	errMigrate := Migrate(ctx)
	if errMigrate != nil {
		slog.Error("could not migrate", "err", errMigrate)
		return 1
	}
	r := chi.NewRouter()
	handler.NewAPI(r, noteService, pool)
	server := httptest.NewServer(r)
	defer server.Close()

	serverURL = server.URL
	return m.Run()
}

func TestEndpoints_Health(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{name: "Success_Health", path: "/healthz", expectedStatus: http.StatusOK},
		{name: "Success_Ready", path: "/readyz", expectedStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+tc.path, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			errBody := resp.Body.Close()
			if errBody != nil {
				slog.Error("Can not close response body", "err", errBody)
			}
		})
	}
}

func TestEndpoint_Notes(t *testing.T) {
	tests := []struct {
		name           string
		request        string
		expectedStatus int
	}{
		{name: "Success_ValidBody", request: `{"text": "test"}`, expectedStatus: http.StatusCreated},
		{name: "BadRequest_InvalidBody", request: `{text: "test"}`, expectedStatus: http.StatusBadRequest},
		{name: "UnprocessableEntity_EmptyBody", request: `{}`, expectedStatus: http.StatusUnprocessableEntity},
		{name: "UnprocessableEntity_EmptyNote", request: `{"text": ""}`, expectedStatus: http.StatusUnprocessableEntity},
		{name: "UnprocessableEntity_WithoutNote", request: `{"note": "test"}`, expectedStatus: http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				serverURL+"/api/v1/notes",
				strings.NewReader(tc.request),
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			postNote, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, tc.expectedStatus, postNote.StatusCode)

			errBody := postNote.Body.Close()
			if errBody != nil {
				slog.Error("Can not close response body", "err", errBody)
			}
		})
	}

	t.Run("Success_GetAll", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, serverURL+"/api/v1/notes", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		errBody := resp.Body.Close()
		if errBody != nil {
			slog.Error("Can not close response body", "err", errBody)
		}
	})
}
