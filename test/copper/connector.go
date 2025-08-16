package copper

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/copper"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

// nolint:gochecknoglobals
var (
	fieldUserEmail = credscanning.Field{
		Name:      "userEmail",
		PathJSON:  "metadata.userEmail",
		SuffixENV: "USER_EMAIL",
	}
)

func GetCopperConnector(ctx context.Context) *copper.Connector {
	filePath := credscanning.LoadPath(providers.Copper)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, fieldUserEmail)

	conn, err := copper.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Copper),
			Metadata: map[string]string{
				"userEmail": reader.Get(fieldUserEmail),
			},
		},
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}
