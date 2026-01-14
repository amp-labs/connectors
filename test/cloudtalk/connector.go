package cloudtalk

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/cloudtalk"
	"github.com/amp-labs/connectors/test/utils"
)

func GetCloudTalkConnector(ctx context.Context) *cloudtalk.Connector {
	filePath := credscanning.LoadPath(providers.CloudTalk)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	// CloudTalk uses Basic Auth.
	client := utils.NewBasicAuthClient(ctx, reader)

	conn, err := cloudtalk.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating cloudtalk connector", "error", err)
	}

	return conn
}
