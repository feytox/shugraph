package test

import (
	"context"
	"fmt"
	"graph-api/internal/app"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

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

	port := 12103
	cnf := &app.Config{
		Port: port,
		DB: app.DBConfig{
			URL: dbURL,
		},
	}

	errMigrate := app.Migrate(ctx, cnf)
	if errMigrate != nil {
		slog.Error("could not migrate", "err", errMigrate)
		return 1
	}

	go func() {
		if err := app.Serve(ctx, cnf); err != nil {
			slog.Error("Ошибка запуска сервера", "err", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	serverURL = fmt.Sprintf("http://localhost:%d", port)
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
			resp, err := http.Get(serverURL + tc.path)

			require.NoError(t, err)
			defer closeBody(resp.Body)

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func closeBody(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		slog.Error("Can not close response body", "err", err)
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
			postNote, err := http.Post(
				serverURL+"/api/v1/notes",
				"application/json",
				strings.NewReader(tc.request))

			require.NoError(t, err)
			defer closeBody(postNote.Body)

			require.Equal(t, tc.expectedStatus, postNote.StatusCode)
		})
	}

	t.Run("Success_GetAll", func(t *testing.T) {
		resp, err := http.Get(serverURL + "/api/v1/notes")

		require.NoError(t, err)
		defer closeBody(resp.Body)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
