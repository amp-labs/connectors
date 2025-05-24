package freshdesk

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/freshdesk"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetFreshdeskConnector(ctx context.Context) *freshdesk.Connector {
	filePath := credscanning.LoadPath(providers.Freshdesk)
	reader := testUtils.MustCreateProvCredJSON(filePath, false)

	conn, err := freshdesk.NewConnector(
		freshdesk.WithClient(ctx, http.DefaultClient, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password)),
		freshdesk.WithWorkspace("pepkarage"),
	)
	if err != nil {
		testUtils.Fail("error creating Freshdesk connector", "error", err)
	}

	return conn
}
