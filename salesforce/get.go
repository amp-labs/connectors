package salesforce

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/jdvr/go-again"
	"github.com/spyzhov/ajson"
)

// get reads data from Salesforce. It handles retries and access token refreshes.
func (s *Connector) get(ctx context.Context, url string) (*ajson.Node, error) {
	var token string

	// Retry will retry the function until it returns a nil error, or a permanent (non-retryable) error.
	return again.Retry[*ajson.Node](ctx, func(ctx context.Context) (*ajson.Node, error) {
		var err error

		// Refresh token if necessary
		if token == "" {
			token, err = s.AccessToken(ctx)
			if err != nil {
				return nil, again.NewPermanentError(err)
			}
		}

		// Add the OAuth header
		authHdr := common.Header{
			Key:   "Authorization",
			Value: "Bearer " + token,
		}

		// Make the request
		d, err := common.GetJson(ctx, s.Client, url, authHdr)
		if err != nil {
			if errors.Is(err, common.ErrApiDisabled) {
				// Not retryable, so return a permanent error
				return nil, again.NewPermanentError(err)
			} else if errors.Is(err, common.ErrAccessToken) {
				// Clear token so that it gets refreshed, then try again.
				token = ""
				return nil, err
			} else if errors.Is(err, common.ErrRetryable) {
				// Retryable error
				return nil, err
			} else {
				// Anything else is a permanent error
				return nil, again.NewPermanentError(err)
			}
		}

		// Success
		return d, nil
	})
}
