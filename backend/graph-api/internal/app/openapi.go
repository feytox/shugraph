package app

import (
	"io"

	"github.com/go-chi/chi/v5"

	"github.com/feytox/shugraph/backend/graph-api/internal/handler"
)

func WriteOpenAPI(w io.Writer) error {
	spec, err := handler.NewAPI(chi.NewRouter(), nil, nil).OpenAPI().YAML()
	if err != nil {
		return err
	}

	_, err = w.Write(spec)

	return err
}
