package sageintacct

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/sageintacct"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetSageIntacctConnector(ctx context.Context) *sageintacct.Connector {
	filePath := credscanning.LoadPath(providers.SageIntacct)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	conn, err := sageintacct.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Sage Intacct connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://api.intacct.com/ia/api/v1/oauth2/authorize",
			TokenURL:  "https://api.intacct.com/ia/api/v1/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return &cfg
}
