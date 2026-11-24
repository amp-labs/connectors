package qargo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/qargo"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// NewConnector builds the connector from the API client's id and secret. Qargo allows 5 token
// requests per hour per tenant, so run the programs a few at a time.
func NewConnector(ctx context.Context) *qargo.Connector {
	filePath := credscanning.LoadPath(providers.Qargo)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	info, err := providers.ReadInfo(providers.Qargo)
	if err != nil {
		utils.Fail("error reading provider info", "error", err)
	}

	config := &clientcredentials.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		TokenURL:     info.Oauth2Opts.TokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	client, err := common.NewOAuthHTTPClient(ctx, common.WithTokenSource(config.TokenSource(ctx)))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := qargo.NewConnector(common.ConnectorParams{AuthenticatedClient: client})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
