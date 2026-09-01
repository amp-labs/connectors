package slackshared

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var signingSecretField = credscanning.Field{
	Name:      "signingSecret",
	PathJSON:  "metadata.signingSecret",
	SuffixENV: "SIGNING_SECRET",
}

func NewConnector(ctx context.Context, provider providers.Provider) *slack.Connector {
	filePath := credscanning.LoadPath(provider)
	reader := utils.MustCreateProvCredJSON(filePath, true, signingSecretField)

	params := common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, makeConfig(provider)),
		Metadata: map[string]string{
			"signingSecret": reader.Get(signingSecretField),
		},
	}

	var (
		conn *slack.Connector
		err  error
	)

	switch provider {
	case providers.Slack:
		conn, err = slack.NewBotConnector(params)
	case providers.SlackUserScope:
		conn, err = slack.NewUserConnector(params)
	default:
		utils.Fail("unknown provider for slack connector", "provider", provider)
	}

	if err != nil {
		utils.Fail("create slack connector", "error: ", err)
	}

	return conn
}

func makeConfig(provider providers.Provider) func(*credscanning.ProviderCredentials) *oauth2.Config {
	tokenUrl := ""
	switch provider {
	case providers.Slack:
		tokenUrl = "https://slack.com/api/oauth.v2.access"
	case providers.SlackUserScope:
		tokenUrl = "https://slack.com/api/oauth.v2.user.access"
	}

	return func(reader *credscanning.ProviderCredentials) *oauth2.Config {

		cfg := &oauth2.Config{
			ClientID:     reader.Get(credscanning.Fields.ClientId),
			ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
			RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://slack.com/oauth/v2/authorize",
				TokenURL:  tokenUrl,
				AuthStyle: oauth2.AuthStyleAutoDetect,
			},
		}

		return cfg
	}
}
