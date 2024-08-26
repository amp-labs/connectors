package smartlead

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	smartlead2 "github.com/amp-labs/connectors/providers/smartlead"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSmartleadConnector(ctx context.Context) *smartlead2.Connector {
	filePath := credscanning.LoadPath(providers.Smartlead)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := smartlead2.NewConnector(
		smartlead2.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
