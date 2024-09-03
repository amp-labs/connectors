package utils

import (
	"github.com/amp-labs/connectors/common/scanning/credscanning"
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
