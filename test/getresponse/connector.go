package getresponse

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetTheGetResponseConnector(ctx context.Context) *aweber.Connector {
	filePath := credscanning.LoadPath(providers.GetResponse)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := aweber.NewConnector(
		aweber.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating GetResponse", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.getresponse.com/oauth2_authorize.html",
			TokenURL:  "https://api.getresponse.com/v3/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}
