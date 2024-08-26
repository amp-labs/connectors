package gong

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/providers/gong"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetGongConnector(ctx context.Context, filePath string) *gong.Connector {
	registry := scanning.NewRegistry()

	readers := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientId",
			KeyName:  utils.ClientId,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientSecret",
			KeyName:  utils.ClientSecret,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.accessToken",
			KeyName:  utils.AccessToken,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.refreshToken",
			KeyName:  utils.RefreshToken,
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.provider",
			KeyName:  utils.Provider,
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
