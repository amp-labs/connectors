package utils

import (
	"fmt"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/utils"
)

// preset values from JSON file from schema equal to response from reference: https://docs.withampersand.com/reference/getconnection
func JSONFileReaders(filePath string) []scanning.Reader {
	schema := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.AccessToken.Token",
			KeyName:  utils.AccessToken,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.RefreshToken.Token",
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

func EnvVarsReaders(prefix string) []scanning.Reader {
	return []scanning.Reader{
		&scanning.EnvReader{
			EnvName: fmt.Sprintf("%sACCESS_TOKEN", prefix),
			KeyName: utils.AccessToken,
		},
		&scanning.EnvReader{
			EnvName: fmt.Sprintf("%sREFRESH_TOKEN", prefix),
			KeyName: utils.RefreshToken,
		},
		&scanning.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_ID", prefix),
			KeyName: utils.ClientId,
		},
		&scanning.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_SECRET", prefix),
			KeyName: utils.ClientSecret,
		},
		&scanning.EnvReader{
			EnvName: fmt.Sprintf("%sWORKSPACE_REF", prefix),
			KeyName: utils.WorkspaceRef,
		},
	}
}

func MustCreateProvCredJSON(filePath string, withAccessToken bool) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewJSONProviderCredentials(filePath, withAccessToken)
	if err != nil {
		Fail("json creds file error", "error", err)
	}

	return reader
}

func MustCreateProvCredENV(providerName string, withAccessToken bool) *credscanning.ProviderCredentials {
	reader, err := credscanning.NewENVProviderCredentials(providerName, withAccessToken)
	if err != nil {
		Fail("environment error", "error", err)
	}

	return reader
}
