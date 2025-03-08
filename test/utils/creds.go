package utils

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"golang.org/x/oauth2"
)

func MustCreateProvCredJSON(filePath string,
	withRequiredAccessToken, withRequiredWorkspace bool,
) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewJSONProviderCredentials(filePath, withRequiredAccessToken, withRequiredWorkspace)
	if err != nil {
		Fail("json creds file error", "error", err)
	}

	return reader
}

// MustCreateProvCredENV can be used by tests supplying variables via environment.
func MustCreateProvCredENV(providerName string,
	withRequiredAccessToken, withRequiredWorkspace bool,
) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewENVProviderCredentials(providerName, withRequiredAccessToken, withRequiredWorkspace)
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
		Fail("failure", "error", err)
	}

	return client
}
