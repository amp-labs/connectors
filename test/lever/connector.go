package lever

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/lever"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *lever.Connector {
	filePath := credscanning.LoadPath(providers.LeverSandbox)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	conn, err := lever.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
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
			AuthURL:   "https://sandbox-lever.auth0.com/authorize",
			TokenURL:  "https://sandbox-lever.auth0.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"offline_access",
			"applications:read:admin",
			"archive_reasons:read:admin",
			"audit_events:read:admin",
			"confidential:access:admin",
			"contact:read:admin",
			"contact:write:admin",
			"feedback:read:admin",
			"feedback:write:admin",
			"feedback_templates:read:admin",
			"feedback_templates:write:admin",
			"files:read:admin",
			"files:write:admin",
			"forms:read:admin",
			"forms:write:admin",
			"form_templates:read:admin",
			"form_templates:write:admin",
			"groups:read:admin",
			"groups:write:admin",
			"interviews:read:admin",
			"interviews:write:admin",
			"notes:read:admin",
			"notes:write:admin",
			"offers:read:admin",
			"opportunities:read:admin",
			"opportunities:write:admin",
			"panels:read:admin",
			"panels:write:admin",
			"permissions:read:admin",
			"permissions:write:admin",
			"postings:read:admin",
			"postings:write:admin",
			"referrals:read:admin",
			"requisitions:read:admin",
			"requisitions:write:admin",
			"requisition_fields:read:admin",
			"requisition_fields:write:admin",
			"resumes:read:admin",
			"roles:read:admin",
			"roles:write:admin",
			"sources:read:admin",
			"stages:read:admin",
			"tags:read:admin",
			"tasks:read:admin",
			"uploads:write:admin",
			"users:read:admin",
			"users:write:admin",
		},
	}

	return &cfg
}
