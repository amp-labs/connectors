package blueshift

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/blueshift"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBlueshiftConnector(ctx context.Context) *blueshift.Connector {
	filePath := credscanning.LoadPath(providers.Blueshift)

	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := blueshift.NewConnector(
		blueshift.WithClient(ctx, http.DefaultClient, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password)),
	)
	if err != nil {
		utils.Fail("error creating asana connector", "error", err)
	}

	return conn
}
