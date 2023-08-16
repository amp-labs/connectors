package salesforce

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/jdvr/go-again"
	"github.com/spyzhov/ajson"
)

func (s *Connector) get(ctx context.Context, url string) (*ajson.Node, error) {
	var token string

	return again.Retry[*ajson.Node](ctx, func(ctx context.Context) (*ajson.Node, error) {
		var err error

		if token == "" {
			token, err = s.AccessToken()
			if err != nil {
				return nil, again.NewPermanentError(err)
			}
		}

		authHdr := common.Header{
			Key:   "Authorization",
			Value: "Bearer " + token,
		}

		d, err := common.GetJson(ctx, s.Client, url, authHdr)
		if err != nil {
			if errors.Is(err, common.ApiDisabled) {
				// Not retryable, so return a permanent error
				return nil, again.NewPermanentError(err)
			} else if errors.Is(err, common.AccessTokenInvalid) {
				// Clear token so that it gets refreshed, then try again.
				token = ""
				return nil, err
			} else if errors.Is(err, common.RetryableError) {
				// Retryable error
				return nil, err
			} else {
				// Anything else is a permanent error
				return nil, again.NewPermanentError(err)
			}
		}

		return d, nil
	})
}
