package front

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/front"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetFrontConnector(ctx context.Context) *front.Connector {
	filePath := credscanning.LoadPath(providers.Front)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := front.NewConnector(
		front.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		testUtils.Fail("error creating Front App connector", "error", err)
	}

	return conn
}
