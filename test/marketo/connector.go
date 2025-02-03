package marketo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/marketo"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2/clientcredentials"
)

func GetMarketoConnector(ctx context.Context) *marketo.Connector {
	reader := getMarketoJSONReader()

	conn, err := marketo.NewConnector(
		marketo.WithClient(ctx, http.DefaultClient, getConfig(reader)),
		marketo.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		marketo.WithModule(marketo.ModuleAssets),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func GetMarketoConnectorLeads(ctx context.Context) *marketo.Connector {
	reader := getMarketoJSONReader()

	conn, err := marketo.NewConnector(
		marketo.WithClient(ctx, http.DefaultClient, getConfig(reader)),
		marketo.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		marketo.WithModule(marketo.ModuleLeads),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *clientcredentials.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	return &clientcredentials.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		TokenURL:     fmt.Sprintf("https://%s.mktorest.com/identity/oauth/token", workspace),
		Scopes:       []string{},
	}
}

func GetMarketoAccessToken() string {
	reader := getMarketoJSONReader()

	return reader.Get(credscanning.Fields.AccessToken)
}

func getMarketoJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Marketo)
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	return reader
}
