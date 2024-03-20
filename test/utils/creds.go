package utils

import (
	"fmt"

	"github.com/amp-labs/connectors/utils"
)

// preset values from JSON file from schema equal to response from reference: https://docs.withampersand.com/reference/getconnection
func JSONFileReaders(filePath string) []utils.Reader {
	schema := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.AccessToken.Token",
			CredKey:  utils.AccessToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.RefreshToken.Token",
			CredKey:  utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientId",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.clientSecret",
			CredKey:  utils.ClientSecret,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerWorkspaceRef",
			CredKey:  utils.WorkspaceRef,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.providerApp.provider",
			CredKey:  utils.Provider,
		},
	}

	return schema
}

func EnvVarsReaders(prefix string) []utils.Reader {
	return []utils.Reader{
		&utils.EnvReader{
			EnvName: fmt.Sprintf("%sACCESS_TOKEN", prefix),
			CredKey: utils.AccessToken,
		},
		&utils.EnvReader{
			EnvName: fmt.Sprintf("%sREFRESH_TOKEN", prefix),
			CredKey: utils.RefreshToken,
		},
		&utils.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_ID", prefix),
			CredKey: utils.ClientId,
		},
		&utils.EnvReader{
			EnvName: fmt.Sprintf("%sCLIENT_SECRET", prefix),
			CredKey: utils.ClientSecret,
		},
		&utils.EnvReader{
			EnvName: fmt.Sprintf("%sWORKSPACE_REF", prefix),
			CredKey: utils.WorkspaceRef,
		},
	}
}
