package getresponse

import (
	"context"

	"github.com/amp-labs/connectors/common"
	provider "github.com/amp-labs/connectors/providers/getresponse"
	"github.com/amp-labs/connectors/test/utils"
)

func GetGetResponseConnector(ctx context.Context) *provider.Connector {
	params := common.ConnectorParams{}

	conn, err := provider.NewConnector(params)
	if err != nil {
		utils.Fail("error creating getresponse connector", "error", err)
	}

	return conn
}
