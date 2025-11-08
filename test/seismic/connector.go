package seismic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/seismic"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *seismic.Connector {
	filePath := credscanning.LoadPath(providers.Seismic)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := seismic.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           "ampersanddemo",
		Module:              providers.ModuleSeismicReporting,
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
			AuthURL:   fmt.Sprintf("https://auth.seismic.com/tenants/%s/connect/authorize", "ampersanddemo"),
			TokenURL:  fmt.Sprintf("https://auth.seismic.com/tenants/%s/connect/token", "ampersanddemo"),
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}
