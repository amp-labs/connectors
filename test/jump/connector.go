package jump

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/jump"
	"github.com/amp-labs/connectors/test/utils"
)

// nolint:gochecknoglobals
var fieldUserID = credscanning.Field{
	Name:      "userId",
	PathJSON:  "metadata.userId",
	SuffixENV: "USER_ID",
}

func GetJumpConnector(ctx context.Context) *jump.Connector {
	filePath := credscanning.LoadPath(providers.Jump)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldUserID)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Jump)

	conn, err := jump.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"userId": reader.Get(fieldUserID),
		},
	})
	if err != nil {
		utils.Fail("error creating Jump connector", "error", err)
	}

	return conn
}
