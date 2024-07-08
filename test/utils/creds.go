package utils

import (
	"fmt"

	"github.com/amp-labs/connectors/common/credsregistry"
	"github.com/amp-labs/connectors/utils"
)

// preset values from JSON file from schema equal to response from reference: https://docs.withampersand.com/reference/getconnection
func JSONFileReaders(filePath string) []credsregistry.Reader {
	schema := []credsregistry.Reader{
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.AccessToken.Token",
			CredKey:  utils.AccessToken,
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.RefreshToken.Token",
			CredKey:  utils.RefreshToken,
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientId",
			CredKey:  utils.ClientId,
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientSecret",
			CredKey:  utils.ClientSecret,
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerWorkspaceRef",
			CredKey:  utils.WorkspaceRef,
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.provider",
			CredKey:  utils.Provider,
		},
	}

	return schema
}

func EnvVarsReaders(prefix string) []credsregistry.Reader {
	return []credsregistry.Reader{
		&credsregistry.EnvReader{
			EnvName: fmt.Sprintf("%sACCESS_TOKEN", prefix),
			CredKey: utils.AccessToken,
		},
		&credsregistry.EnvReader{
			EnvName: fmt.Sprintf("%sREFRESH_TOKEN", prefix),
			CredKey: utils.RefreshToken,
		},
		&credsregistry.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_ID", prefix),
			CredKey: utils.ClientId,
		},
		&credsregistry.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_SECRET", prefix),
			CredKey: utils.ClientSecret,
		},
		&credsregistry.EnvReader{
			EnvName: fmt.Sprintf("%sWORKSPACE_REF", prefix),
			CredKey: utils.WorkspaceRef,
		},
	}
}

func MustCreateProvCredJSON(filePath string, withAccessToken bool) *credsregistry.ProviderCredentials {
	reader, err := credsregistry.NewJSONProviderCredentials(filePath, withAccessToken)
	if err != nil {
		Fail("json creds file error", "error", err)
	}

	return reader
}

func MustCreateProvCredENV(providerName string, withAccessToken bool) *credsregistry.ProviderCredentials {
	reader, err := credsregistry.NewENVProviderCredentials(providerName, withAccessToken)
	if err != nil {
		Fail("environment error", "error", err)
	}

	return reader
}
