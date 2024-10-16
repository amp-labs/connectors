package utils

import (
	"github.com/amp-labs/connectors/common/scanning/credscanning"
)

<<<<<<< HEAD
=======
// preset values from JSON file from schema equal to response from reference: https://docs.withampersand.com/reference/getconnection
func JSONFileReaders(filePath string) []scanning.Reader {
	schema := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.AccessToken.token",
			KeyName:  utils.AccessToken,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.RefreshToken.token",
			KeyName:  utils.RefreshToken,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientId",
			KeyName:  utils.ClientId,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientSecret",
			KeyName:  utils.ClientSecret,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerWorkspaceRef",
			KeyName:  utils.WorkspaceRef,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.provider",
			KeyName:  utils.Provider,
		},
	}

	return schema
}

>>>>>>> 1076d84 (fix sf subscribe related methods)
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
