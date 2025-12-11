package blackbaud

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/blackbaud"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldSubscriptionKey = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "bbApiSubscriptionKey",
	PathJSON:  "metadata.bbApiSubscriptionKey",
	SuffixENV: "BB_API_SUBSCRIPTION_KEY",
}

func GetBlackbaudConnector(ctx context.Context) *blackbaud.Connector {
	filePath := credscanning.LoadPath(providers.Blackbaud)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldSubscriptionKey)

	conn, err := blackbaud.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Metadata: map[string]string{
				"bbApiSubscriptionKey": reader.Get(fieldSubscriptionKey),
			},
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.blackbaud.com/oauth/authorize",
			TokenURL:  "https://oauth2.sky.blackbaud.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}
