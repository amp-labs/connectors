package meta

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	metaconn "github.com/amp-labs/connectors/providers/meta"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var ( //nolint:gochecknoglobals
	fieldWhatsAppAccountID = credscanning.Field{
		Name:      "whatsappAccountId",
		PathJSON:  "metadata.whatsappAccountId",
		SuffixENV: "WHATSAPP_ACCOUNT_ID",
	}
	fieldWhatsAppPhoneNumberID = credscanning.Field{
		Name:      "whatsappPhoneNumberId",
		PathJSON:  "metadata.whatsappPhoneNumberId",
		SuffixENV: "WHATSAPP_PHONE_NUMBER_ID",
	}
)

func GetWhatsAppConnector(ctx context.Context) *metaconn.Connector {
	filePath := credscanning.LoadPath(providers.Meta)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldWhatsAppAccountID, fieldWhatsAppPhoneNumberID)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := metaconn.NewConnector(common.ConnectorParams{
		Module:              providers.ModuleMetaWhatsApp,
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"whatsappAccountId":     reader.Get(fieldWhatsAppAccountID),
			"whatsappPhoneNumberId": reader.Get(fieldWhatsAppPhoneNumberID),
		},
	})
	if err != nil {
		utils.Fail("error creating meta whatsapp connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://www.facebook.com/v25.0/dialog/oauth",
			TokenURL:  "https://graph.facebook.com/v25.0/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"whatsapp_business_messaging",
			"whatsapp_business_management",
		},
	}
}
