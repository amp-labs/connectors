package brevo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/brevo"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBrevoConnector(ctx context.Context) *brevo.Connector {
	filePath := credscanning.LoadPath(providers.Brevo)

	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "api-key", reader.Get(credscanning.Fields.ApiKey))

	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := brevo.NewConnector(
		parameters.Connector{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		utils.Fail("error creating brevo connector", "error", err)
	}

	return conn
}
