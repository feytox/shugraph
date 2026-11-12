package graphapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi/oapi"
)

func (c *Client) LoginEvent(ctx context.Context, eventID string, payload oapi.LoginPayload) error {
	resp, err := c.api.LoginEventWithResponse(ctx, eventID, payload)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrGraphAPIUnavailable, err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("%w: unexpected status %d", domain.ErrGraphAPIUnavailable, resp.StatusCode())
	}
	return nil
}

func (c *Client) TransferEvent(ctx context.Context, eventID string, payload oapi.TransferPayload) error {
	resp, err := c.api.TransferEventWithResponse(ctx, eventID, payload)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrGraphAPIUnavailable, err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("%w: unexpected status %d", domain.ErrGraphAPIUnavailable, resp.StatusCode())
	}
	return nil
}
