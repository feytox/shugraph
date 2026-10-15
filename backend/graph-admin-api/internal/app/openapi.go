package app

import (
	"io"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/handler"
)

func WriteOpenAPI(w io.Writer) error {
	spec, err := handler.OpenAPISpec()
	if err != nil {
		return err
	}

	_, err = w.Write(spec)
	return err
}
