package hubspot

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

// GetHubspotConnector returns a Hubspot connector.
func GetHubspotConnector(ctx context.Context) *hubspot.Connector {
	reader := CredsReader()

	conn, err := hubspot.NewConnector(
		hubspot.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		hubspot.WithModule(hubspot.ModuleCRM))
	if err != nil {
		utils.Fail("error creating hubspot connector", "error", err)
	}

	return conn
}

func CredsReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Hubspot)
	return utils.MustCreateProvCredJSON(filePath, true, false)
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.hubspot.com/oauth/authorize",
			TokenURL:  "https://api.hubapi.com/oauth/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"crm.objects.contacts.read",
			"crm.objects.contacts.write",
			"crm.objects.deals.read",
			"crm.objects.line_items.read",
			"oauth",
			"crm.objects.companies.read",
			"tickets",
		},
	}

	return cfg
}
