package graphapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi/oapi"
)

type Client struct {
	api *oapi.ClientWithResponses
}

func New(baseURL string, httpClient *http.Client) (*Client, error) {
	api, err := oapi.NewClientWithResponses(baseURL,
		oapi.WithHTTPClient(httpClient),
		oapi.WithRequestEditorFn(forwardRequestID),
	)
	if err != nil {
		return nil, fmt.Errorf("create graph-api client: %w", err)
	}

	return &Client{api: api}, nil
}

func (c *Client) CreateNote(ctx context.Context, text string) (string, error) {
	resp, err := c.api.CreateNoteWithResponse(ctx, oapi.CreateNoteJSONRequestBody{Text: text})
	if err != nil {
		return "", fmt.Errorf("%w: %w", domain.ErrGraphAPIUnavailable, err)
	}

	if resp.JSON201 == nil {
		return "", fmt.Errorf("%w: unexpected status %d", domain.ErrGraphAPIUnavailable, resp.StatusCode())
	}

	return resp.JSON201.Id, nil
}

func forwardRequestID(ctx context.Context, req *http.Request) error {
	if id := middleware.GetReqID(ctx); id != "" {
		req.Header.Set(middleware.RequestIDHeader, id)
	}

	return nil
}
