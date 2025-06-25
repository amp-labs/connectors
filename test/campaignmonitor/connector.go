package campaignmonitor

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/campaignmonitor"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

// nolint:gochecknoglobals
var (
	fieldClientID = credscanning.Field{
		Name:      "clientId",
		PathJSON:  "metadata.clientId",
		SuffixENV: "CLIENT_ID",
	}
)

func GetCampaignMonitorConnector(ctx context.Context) *campaignmonitor.Connector {
	filePath := credscanning.LoadPath(providers.CampaignMonitor)
	reader := utils.MustCreateProvCredJSON(filePath, true, false, fieldClientID)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := campaignmonitor.NewConnector(common.Parameters{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"clientId": reader.Get(fieldClientID),
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
			TokenURL:  "https://api.createsend.com",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return &cfg
}
