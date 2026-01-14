package shopify

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/shopify"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetShopifyConnector(ctx context.Context) *shopify.Connector {
	filePath := credscanning.LoadPath(providers.Shopify)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	workspace := reader.Get(credscanning.Fields.Workspace)

	// Shopify uses OAuth2 with a custom header (X-Shopify-Access-Token)
	client := utils.NewOauth2ClientForProvider(ctx, providers.Shopify, reader, getOAuthConfig(workspace))

	conn, err := shopify.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           workspace,
	})
	if err != nil {
		utils.Fail("error creating Shopify connector", "error", err)
	}

	return conn
}

func getOAuthConfig(workspace string) func(*credscanning.ProviderCredentials) *oauth2.Config {
	return func(reader *credscanning.ProviderCredentials) *oauth2.Config {
		return &oauth2.Config{
			ClientID:     reader.Get(credscanning.Fields.ClientId),
			ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
			RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://" + workspace + ".myshopify.com/admin/oauth/authorize",
				TokenURL:  "https://" + workspace + ".myshopify.com/admin/oauth/access_token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{
				"read_customers", "write_customers",
				"read_products", "write_products",
				"read_orders", "write_orders",
			},
		}
	}
}
