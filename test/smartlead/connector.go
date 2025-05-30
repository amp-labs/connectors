package smartlead

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/smartlead"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSmartleadConnector(ctx context.Context) *smartlead.Connector {
	filePath := credscanning.LoadPath(providers.Smartlead)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := smartlead.NewConnector(
		smartlead.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
