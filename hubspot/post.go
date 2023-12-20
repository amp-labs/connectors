package hubspot

import (
	"context"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// post writes data to Hubspot. It handles retries and access token refreshes.
func (c *Connector) post(ctx context.Context, url string, body any) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Post(ctx, url, body)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrAccessToken):
			// TODO: Retry
			slog.Warn("Access token invalid, retrying", "error", err)

			// TODO: Retry
			return nil, err
		case errors.Is(err, common.ErrRetryable):
			// TODO: Retry
			return nil, err
		case errors.Is(err, common.ErrApiDisabled):
			fallthrough
		case errors.Is(err, common.ErrForbidden):
			fallthrough
		default:
			// Anything else is a permanent error
			return nil, err
		}
	}

	// Success
	return rsp, nil
}
