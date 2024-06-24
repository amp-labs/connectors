package gong

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/providers/gong"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetGongConnector(ctx context.Context, filePath string) *gong.Connector {
	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientId",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientSecret",
			CredKey:  utils.ClientSecret,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.accessToken",
			CredKey:  utils.AccessToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.refreshToken",
			CredKey:  utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.provider",
			CredKey:  utils.Provider,
		},
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.GongOAuthConfigFromRegistry(registry)
	tok := utils.GongOauthTokenFromRegistry(registry)

	conn, err := gong.NewConnector(
		gong.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}
