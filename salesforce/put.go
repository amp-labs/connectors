package salesforce

import (
	"context"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) put(ctx context.Context, url string, body []byte) ([]byte, error) {
	resBody, err := c.Client.PutCSV(ctx, url, body)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrAccessToken):
			// Retryable, so just log and retry
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

	return resBody, nil
}
