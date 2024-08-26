package instantly

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/test/utils"
)

func GetInstantlyConnector(ctx context.Context) *instantly.Connector {
	filePath := credscanning.LoadPath(providers.Instantly)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := instantly.NewConnector(
		instantly.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
