package salesforce

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/jdvr/go-again"
	"github.com/spyzhov/ajson"
)

// get reads data from Salesforce. It handles retries and access token refreshes.
func (c *Connector) get(ctx context.Context, url string) (*ajson.Node, error) {
	// Retry will retry the function until it returns a nil error, or a permanent (non-retryable) error.
	return again.Retry[*ajson.Node](ctx, func(ctx context.Context) (*ajson.Node, error) {
		// Make the request
		node, err := c.Client.Get(ctx, url)
		if err != nil {
			switch {
			case errors.Is(err, common.ErrApiDisabled):
				// Not retryable, so return a permanent error
				return nil, again.NewPermanentError(err)
			case errors.Is(err, common.ErrForbidden):
				// Not retryable, so return a permanent error
				return nil, again.NewPermanentError(err)
			case errors.Is(err, common.ErrAccessToken):
				// Retryable, so just log and retry
				slog.Warn("Access token invalid, retrying", "error", err)

				return nil, err
			case errors.Is(err, common.ErrRetryable):
				return nil, err
			default:
				// Anything else is a permanent error
				return nil, again.NewPermanentError(err)
			}
		}

		// Success
		return node, nil
	})
}

func (c *Connector) interpretError(rsp *http.Response, body []byte) error {
	return common.InterpretError(rsp, body)
}
