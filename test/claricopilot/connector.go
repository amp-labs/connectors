package claricopilot

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/claricopilot"
	"github.com/amp-labs/connectors/test/utils"
)

func GetClariCopilotConnector(ctx context.Context) *claricopilot.Connector {
	filePath := credscanning.LoadPath(providers.ClariCopilot)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := NewAPIKeyClient(ctx, reader)

	conn, err := claricopilot.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating ClariCopilot connector", "error", err)
	}

	return conn
}

func NewAPIKeyClient(
	ctx context.Context, reader *credscanning.ProviderCredentials,
) common.AuthenticatedHTTPClient {

	var (
		headerNameKey        = "secret-key"
		headerValueKey       = "this-is-a-secret"
		headerSecretKey      = "secret-key"
		headerValueSecretKey = "this-is-a-secret"
	)

	client, err := claricopilot.NewClariCopilotAuthHTTPClient(ctx, headerNameKey, headerValueKey, headerSecretKey, headerValueSecretKey)
	if err != nil {
		utils.Fail("error creating ClariCopilot client", "error", err)
	}

	return client
}
