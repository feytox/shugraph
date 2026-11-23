package handler

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPIUpToDate(t *testing.T) {
	generated, err := OpenAPISpec()
	require.NoError(t, err)

	committed, err := os.ReadFile("../../openapi.yaml")
	require.NoError(t, err)

	require.Equal(t, string(generated), strings.ReplaceAll(string(committed), "\r\n", "\n"), "openapi.yaml is stale: run make openapi")
}
