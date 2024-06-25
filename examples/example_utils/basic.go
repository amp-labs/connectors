package example_utils

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type BasicAuthOptions struct {
	User string
	Pass string
}

func CreateBasicAuthClient(ctx context.Context, info *providers.ProviderInfo, opts BasicAuthOptions) common.AuthenticatedHTTPClient {
	// Create the authenticated HTTP client.
	httpClient, err := info.NewClient(ctx, &providers.NewClientParams{
		// If you set this to true, the client will log all requests and responses.
		// Be careful with this in production, as it may expose sensitive data.
		Debug: *debug,

		// If you have your own HTTP client, you can use it here.
		Client: http.DefaultClient,

		// BasicCreds represents the basic authentication credentials.
		BasicCreds: &providers.BasicParams{
			User: opts.User,
			Pass: opts.Pass,
		},
	})
	if err != nil {
		panic(err)
	}

	return httpClient
}
