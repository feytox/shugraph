package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractRoutes(t *testing.T) {
	const groupPath = "/api/v1/groups/00000000-0000-0000-0000-000000000001"

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/groups", ""},
		{http.MethodGet, groupPath, ""},
		{http.MethodGet, groupPath + "/graph", ""},
		{http.MethodPatch, groupPath + "/solution", `{"status":"approved","decision_comment":"Confirmed"}`},
	}

	handler := New(Deps{Log: slog.New(slog.DiscardHandler)})

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusNotImplemented, rec.Code, rec.Body.String())
		})
	}
}

func TestOpenAPIUpToDate(t *testing.T) {
	generated, err := OpenAPISpec()
	require.NoError(t, err)

	committed, err := os.ReadFile("../../openapi.yaml")
	require.NoError(t, err)

	require.Equal(t, string(generated), strings.ReplaceAll(string(committed), "\r\n", "\n"), "openapi.yaml is stale: run make openapi")
}
