package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
		{http.MethodPatch, groupPath + "/solution", `{"approved":true,"decision_comment":"Confirmed"}`},
		{http.MethodPut, "/api/v1/events/logins/login-1", `{
			"occurred_at":"2026-10-26T12:00:00Z",
			"account_id":"account-1","client_id":"client-1",
			"device_id":"device-1","address_id":"127.0.0.1"
		}`},
		{http.MethodPut, "/api/v1/events/transfers/transfer-1", `{
			"occurred_at":"2026-10-26T12:00:00Z",
			"from_account_id":"account-1","from_client_id":"client-1",
			"to_account_id":"account-2","to_client_id":"client-2","amount":100
		}`},
	}

	handler := New(Deps{Log: slog.New(slog.DiscardHandler)})

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("expected contract stub status 501, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestOpenAPIUpToDate(t *testing.T) {
	generated, err := OpenAPISpec()
	if err != nil {
		t.Fatal(err)
	}

	committed, err := os.ReadFile("../../openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if strings.ReplaceAll(string(committed), "\r\n", "\n") != string(generated) {
		t.Fatal("openapi.yaml is stale: run make openapi")
	}
}
