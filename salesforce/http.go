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

	ctx = context.WithValue(ctx, oauth2.HTTPClient, params.client)

	if params.tokenSource != nil {
		return oauth2.NewClient(ctx, params.tokenSource)
	} else {
		return oauth2.NewClient(ctx, params.config.TokenSource(ctx, params.token))
	}
}
