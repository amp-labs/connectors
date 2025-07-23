package utils

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

func MustCreateProvCredJSON(filePath string,
	withRequiredAccessToken bool,
	customFields ...credscanning.Field,
) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewJSONProviderCredentials(
		filePath, withRequiredAccessToken, customFields...,
	)
	if err != nil {
		Fail("json creds file error", "error", err)
	}

	return reader
}

// MustCreateProvCredENV can be used by tests supplying variables via environment.
func MustCreateProvCredENV(providerName string,
	withRequiredAccessToken bool,
	customFields ...credscanning.Field,
) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewENVProviderCredentials(providerName, withRequiredAccessToken, customFields...)
	if err != nil {
		Fail("environment error", "error", err)
	}

	return reader
}

func NewOauth2Client(
	ctx context.Context,
	reader *credscanning.ProviderCredentials,
	configProvider func(*credscanning.ProviderCredentials) *oauth2.Config,
) common.AuthenticatedHTTPClient {
	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(configProvider(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		Fail("error creating oauth", "error", err)
	}

	return client
}

func NewAPIKeyClient(
	ctx context.Context, reader *credscanning.ProviderCredentials, provider providers.Provider,
) common.AuthenticatedHTTPClient {
	providerInfo, err := providers.ReadInfo(provider)
	if err != nil {
		Fail("error reading provider info", "error", err)
	}

	headerName, headerValue, err := providerInfo.GetApiKeyHeader(reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		Fail("error getting API key header", "error", err)
	}

	client, err := common.NewApiKeyHeaderAuthHTTPClient(
		ctx, headerName, headerValue)
	if err != nil {
		Fail("error creating API key client", "error", err)
	}

	return client
}

func NewBasicAuthClient(
	ctx context.Context, reader *credscanning.ProviderCredentials,
) common.AuthenticatedHTTPClient {
	username := reader.Get(credscanning.Fields.Username)
	password := reader.Get(credscanning.Fields.Password)

	client, err := common.NewBasicAuthHTTPClient(ctx, username, password)
	if err != nil {
		Fail("error creating basic auth client", "error", err)
	}

	return client
}

func NewCustomAuthClient(
	ctx context.Context,
	reader *credscanning.ProviderCredentials,
	provider providers.Provider,
	fields ...credscanning.Field,
) common.AuthenticatedHTTPClient {
	providerInfo, err := providers.ReadInfo(provider)
	if err != nil {
		Fail("error reading provider info", "error", err)
	}

	vals := make(map[string]string)

	for _, field := range fields {
		val := reader.Get(field)
		if val == "" {
			Fail("missing custom auth field value", "field", field.Name)
		}

		vals[field.Name] = val
	}

	client, err := providerInfo.NewClient(ctx, &providers.NewClientParams{
		CustomCreds: &providers.CustomAuthParams{
			Values: vals,
		},
	})

	if err != nil {
		Fail("error creating custom auth client", "error", err)
	}

	return client
}
