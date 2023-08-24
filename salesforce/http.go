package salesforce

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

func newHTTPClient(ctx context.Context, params *sfParams) common.HTTPClient { //nolint:ireturn
	if params.client == nil {
		params.client = http.DefaultClient
	}

	// This is how the key refresher accepts a custom http client
	ctx = context.WithValue(ctx, oauth2.HTTPClient, params.client)

	// Returns a new client which automatically refreshes the access token
	// whenever the current one expires.
	if params.tokenSource != nil {
		return oauth2.NewClient(ctx, params.tokenSource)
	} else {
		return oauth2.NewClient(ctx, params.config.TokenSource(ctx, params.token))
	}
}
