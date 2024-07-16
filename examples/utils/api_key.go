package utils

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type ApiKeyOptions struct {
	ApiKey string
}

func CreateApiKeyClient(ctx context.Context, info *providers.ProviderInfo, opts ApiKeyOptions) common.AuthenticatedHTTPClient {
	// Create the authenticated HTTP client.
	httpClient, err := info.NewClient(ctx, &providers.NewClientParams{
		// If you set this to true, the client will log all requests and responses.
		// Be careful with this in production, as it may expose sensitive data.
		Debug: *debug,
		// If you have your own HTTP client, you can use it here.
		Client: http.DefaultClient,
		ApiKey: opts.ApiKey,
	})
	if err != nil {
		panic(err)
	}

	return httpClient
}
