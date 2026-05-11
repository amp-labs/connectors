package ramp

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/ramp"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetRampConnector(ctx context.Context) *ramp.Connector {
	filePath := credscanning.LoadPath(providers.RampDemo)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := ramp.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
	})
	if err != nil {
		utils.Fail("error creating ramp connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://api.ramp.com/v1/authorize",
			TokenURL:  "https://api.ramp.com/developer/v1/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		Scopes: []string{
			"transactions:read",
			"users:read",
			"users:write",
			"cards:read",
			"departments:read",
			"departments:write",
			"locations:read",
			"locations:write",
			"vendors:read",
			"vendors:write",
			"reimbursements:read",
			"limits:read",
			"bills:read",
			"bills:write",
		},
	}
}
