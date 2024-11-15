package dynamicscrm

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/klaviyo"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetKlaviyoConnector(ctx context.Context) *klaviyo.Connector {
	filePath := credscanning.LoadPath(providers.Klaviyo)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := klaviyo.NewConnector(
		klaviyo.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Klaviyo connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://www.klaviyo.com/oauth/authorize",
			TokenURL:  "https://a.klaviyo.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"accounts:write",
			"flows:write",
			"segments:write",
			"segments:read",
			"data-privacy:read",
			"images:read",
			"metrics:read",
			"profiles:read",
			"catalogs:read",
			"push-tokens:read",
			"push-tokens:write",
			"templates:read",
			"tags:read",
			"templates:write",
			"coupons:read",
			"events:read",
			"images:write",
			"lists:read",
			"catalogs:write",
			"coupon-codes:read",
			"campaigns:write",
			"flows:read",
			"subscriptions:read",
			"conversations:write",
			"coupons:write",
			"campaigns:read",
			"lists:write",
			"data-privacy:write",
			"events:write",
			"subscriptions:write",
			"profiles:write",
			"tags:write",
			"accounts:read",
			"conversations:read",
			"coupon-codes:write",
			"metrics:write",
		},
	}
}
