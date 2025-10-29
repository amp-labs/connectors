package paypal

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/paypal"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2/clientcredentials"
)

func GetPayPalConnector(ctx context.Context) *paypal.Connector {
	filePath := credscanning.LoadPath(providers.PayPalSandBox)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	cc := getConfig(reader)
	tt := cc.TokenSource(ctx)

	client, err := common.NewOAuthHTTPClient(ctx, common.WithTokenSource(tt))
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := paypal.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		TokenURL:     "https://api-m.sandbox.paypal.com/v1/oauth2/token",
	}
}
