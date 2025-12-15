package getresponse

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGetResponseConnector(ctx context.Context) *getresponse.Connector {
	filePath := credscanning.LoadPath(providers.GetResponse)

	reader := utils.MustCreateProvCredJSON(filePath, true)

	params := common.ConnectorParams{

		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
	}

	conn, err := getresponse.NewConnector(params)
	if err != nil {
		utils.Fail("error creating getresponse connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.getresponse.com/oauth2_authorize.html",
			TokenURL:  "https://api.getresponse.com/v3/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}
