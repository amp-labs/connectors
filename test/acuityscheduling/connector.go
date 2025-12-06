package acuityscheduling

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	acuityScheduling "github.com/amp-labs/connectors/providers/acuityscheduling"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetAcuitySchedulingConnector(ctx context.Context) *acuityScheduling.Connector {
	filePath := credscanning.LoadPath(providers.AcuityScheduling)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := acuityScheduling.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
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

		RedirectURL: "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://acuityscheduling.com/oauth2/authorize",
			TokenURL:  "https://acuityscheduling.com/oauth2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}
