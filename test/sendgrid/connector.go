package sendgrid

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/sendgrid"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSendGridConnector(ctx context.Context) *sendgrid.Connector {
	filePath := credscanning.LoadPath(providers.SendGrid)

	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(
		ctx,
		"Authorization",
		"Bearer "+reader.Get(credscanning.Fields.ApiKey),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := sendgrid.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating SendGrid connector", "error", err)
	}

	return conn
}
