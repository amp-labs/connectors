package meta

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/meta"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var (
	fieldAdAccountId = credscanning.Field{
		Name:      "adAccountId",
		PathJSON:  "metadata.adAccountId",
		SuffixENV: "AD_ACCOUNT_ID",
	}
	fieldBusinessId = credscanning.Field{
		Name:      "businessId",
		PathJSON:  "metadata.businessId",
		SuffixENV: "BUSINESS_ID",
	}
)

func GetFacebookConnector(ctx context.Context) *meta.Connector {
	return GetConnector(ctx, providers.ModuleFacebook)
}

func GetConnector(ctx context.Context, moduleID common.ModuleID) *meta.Connector {
	filePath := credscanning.LoadPath(providers.Meta)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldAdAccountId, fieldBusinessId)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	conn, err := meta.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Module:              moduleID,
		Metadata: map[string]string{
			"adAccountId": reader.Get(fieldAdAccountId),
			"businessId":  reader.Get(fieldBusinessId),
		},
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
			AuthURL:   "https://www.facebook.com/v19.0/dialog/oauth",
			TokenURL:  "https://graph.facebook.com/v19.0/oauth/access_token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{
			"ads_management",
			"ads_read",
			"attribution_read",
			"business_management",
			"commerce_account_read_reports",
			"commerce_account_read_settings",
			"email",
			"instagram_branded_content_ads_brand",
			"instagram_branded_content_brand",
			"instagram_branded_content_creator",
			"instagram_content_publish",
			"instagram_manage_comments",
			"instagram_manage_events",
			"instagram_manage_messages",
			"instagram_manage_upcoming_events",
			"manage_fundraisers",
			"pages_manage_ads",
			"pages_manage_cta",
			"pages_manage_instant_articles",
			"pages_manage_engagement",
			"pages_manage_metadata",
			"pages_manage_posts",
			"pages_messaging",
			"pages_read_engagement",
			"pages_read_user_content",
			"pages_show_list",
			"private_computation_access",
			"public_profile",
			"publish_video",
			"threads_business_basic",
			"whatsapp_business_manage_events",
			"whatsapp_business_management",
			"whatsapp_business_messaging",
		},
	}

	return &cfg
}
