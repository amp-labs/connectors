package kaseyavsax

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/kaseyavsax"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func NewConnector(ctx context.Context) *kaseyavsax.Connector {
	filePath := credscanning.LoadPath(providers.KaseyaVSAX)
	reader := testUtils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password))
	if err != nil {
		testUtils.Fail("error creating KaseyaVSAx connector", "error", err)
	}

	conn, err := kaseyavsax.NewConnector(
		common.ConnectorParams{
			Workspace:           "agenticai.vsax.net",
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		testUtils.Fail("error creating KaseyaVSAX connector", "error", err)
	}

	return conn
}
