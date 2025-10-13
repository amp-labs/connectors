package highlevelwhitelabel

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/highlevelwhitelabel"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldLocationId = credscanning.Field{
	Name:      "locationId",
	PathJSON:  "metadata.locationId",
	SuffixENV: "LOCATION_ID",
}

func GetHighLevelWhiteLabelConnector(ctx context.Context) *highlevelwhitelabel.Connector {
	filePath := credscanning.LoadPath(providers.HighLevelWhiteLabel)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldLocationId)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := highlevelwhitelabel.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"locationId": reader.Get(fieldLocationId),
		},
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
			AuthURL:   "https://marketplace.leadconnectorhq.com/oauth/chooselocation",
			TokenURL:  "https://services.leadconnectorhq.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"businesses.readonly",
			"businesses.write",
			"calendars.readonly",
			"calendars.write",
			"forms.write",
			"forms.readonly",
			"users.readonly",
			"users.write",
			"locations.write",
			"locations.readonly",
		},
	}

	return &cfg
}
