package gusto

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gusto"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

//nolint:gochecknoglobals
var fieldCompanyID = credscanning.Field{
	Name:      "companyId",
	PathJSON:  "metadata.companyId",
	SuffixENV: "COMPANY_ID",
}

func GetConnector(ctx context.Context) *gusto.Connector {
	filePath := credscanning.LoadPath(providers.GustoDemo)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldCompanyID)

	conn, err := gusto.NewDemoConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Metadata: map[string]string{
				"companyId": reader.Get(fieldCompanyID),
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
			AuthURL:   "https://api.gusto-demo.com/oauth/authorize",
			TokenURL:  "https://api.gusto-demo.com/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}
