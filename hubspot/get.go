package hubspot

import (
	"context"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

// get reads data from Hubspot. It handles retries and access token refreshes.
func (c *Connector) get(ctx context.Context, url string) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Get(ctx, url)
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
