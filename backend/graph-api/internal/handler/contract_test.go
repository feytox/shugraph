package handler

import (
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestOpenAPIUpToDate(t *testing.T) {
	generated, err := NewAPI(chi.NewRouter(), nil, nil).OpenAPI().YAML()
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
