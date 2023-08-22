package salesforce

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

func newHTTPClient(params *sfParams) common.HTTPClient { //nolint:ireturn
	if params.client == nil {
		params.client = http.DefaultClient
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, params.client)

	if params.tokenSource != nil {
		return oauth2.NewClient(ctx, params.tokenSource)
	} else {
		return oauth2.NewClient(ctx, params.config.TokenSource(ctx, params.token))
	}
}

type client struct {
	c common.HTTPClient
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	slog.Info("making request", "req", req)

	return c.c.Do(req)
}

func (c *client) CloseIdleConnections() {
	c.c.CloseIdleConnections()
}

func wrapClient(c common.HTTPClient) common.HTTPClient { //nolint:ireturn
	return &client{c}
}
