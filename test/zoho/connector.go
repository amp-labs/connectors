package zoho

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetZohoConnector(ctx context.Context, module common.ModuleID) *zoho.Connector {
	return GetZohoConnectorWithMetadata(ctx, module, nil)
}

// GetZohoConnectorNoRefresh builds a connector that uses the creds access token
// verbatim: it stamps a far-future expiry so the OAuth client never
// pre-emptively refreshes (credscanning.GetOauthToken otherwise back-dates the
// expiry, forcing an immediate refresh that would swap in a token minted from
// the refresh token — with whatever scopes that grant carries). Use this when
// the access token in the creds file was minted with specific scopes that must
// be sent as-is.
func GetZohoConnectorNoRefresh(
	ctx context.Context, module common.ModuleID, metadata map[string]string,
) *zoho.Connector {
	filePath := credscanning.LoadPath(providers.Zoho)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	token := reader.GetOauthToken()
	token.Expiry = time.Now().Add(365 * 24 * time.Hour) //nolint:mnd // far future: never refresh

	opts := []zoho.Option{
		zoho.WithClient(ctx, http.DefaultClient, getConfig(reader), token),
		zoho.WithModule(module),
	}
	if metadata != nil {
		opts = append(opts, zoho.WithMetadata(metadata))
	}

	conn, err := zoho.NewConnector(opts...)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	fmt.Println("Module: ", module)

	return conn
}

// GetZohoConnectorWithMetadata builds a Zoho connector and passes the given
// connection metadata (e.g. the Zoho Mail webhook signing secret under
// "zohoMailWebhookSecret", or "zohoMailAccountId" for account-scoped calls).
func GetZohoConnectorWithMetadata(
	ctx context.Context, module common.ModuleID, metadata map[string]string,
) *zoho.Connector {
	filePath := credscanning.LoadPath(providers.Zoho)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	opts := []zoho.Option{
		zoho.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		zoho.WithModule(module),
	}
	if metadata != nil {
		opts = append(opts, zoho.WithMetadata(metadata))
	}

	conn, err := zoho.NewConnector(opts...)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	fmt.Println("Module: ", module)

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:  "https://accounts.zoho.com/oauth/v2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{
			"ZohoCRM.modules.ALL",
			"ZohoCRM.settings.ALL",
			"ZohoCRM.notifications.ALL",
		},
	}

	return &cfg
}
