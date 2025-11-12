package linkedin

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/linkedin"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldAdAccountId = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "adAccountId",
	PathJSON:  "metadata.adAccountId",
	SuffixENV: "AD_ACCOUNT_ID",
}

func GetPlatformConnector(ctx context.Context) *linkedin.Connector {
	return getConnector(ctx, providers.ModuleLinkedInPlatform)
}

func GetAdsConnector(ctx context.Context) *linkedin.Connector {
	return getConnector(ctx, providers.ModuleLinkedInAds)
}

func getConnector(ctx context.Context, moduleID common.ModuleID) *linkedin.Connector {
	filePath := credscanning.LoadPath(providers.LinkedIn)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldAdAccountId)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	metadata := map[string]string{}

	if moduleID == providers.ModuleLinkedInAds {
		metadata["adAccountId"] = reader.Get(fieldAdAccountId)
	}

	conn, err := linkedin.NewConnector(common.ConnectorParams{
		Module:              moduleID,
		AuthenticatedClient: client,
		Metadata:            metadata,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:  "https://www.linkedin.com/oauth/v2/accessToken",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"rw_ads", "r_ads",
		},
	}

	return &cfg
}
