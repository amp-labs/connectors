package pipedrive

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetPipedriveConnector(ctx context.Context) *pipedrive.Connector {
	filePath := credscanning.LoadPath(providers.Pipedrive)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := pipedrive.NewConnector(
		pipedrive.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		pipedrive.WithModule(providers.ModulePipedriveCRM),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth/wewe",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://oauth.pipedrive.com/oauth/authorize",
			TokenURL:  "https://oauth.pipedrive.com/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{},
	}

	return &cfg
}
