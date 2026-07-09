package mailgun

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

func GetMailgunConnector(ctx context.Context) *mailgun.Connector {
	filePath := credscanning.LoadPath(providers.Mailgun)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewBasicAuthClient(ctx, reader)

	conn, err := mailgun.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating mailgun connector", "error", err)
	}

	return conn
}
