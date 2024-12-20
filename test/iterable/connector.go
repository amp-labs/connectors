package iterable

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/iterable"
	"github.com/amp-labs/connectors/test/utils"
)

func GetIterableConnector(ctx context.Context) *iterable.Connector {
	filePath := credscanning.LoadPath(providers.Iterable)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := iterable.NewConnector(
		iterable.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating Iterable connector", "error", err)
	}

	return conn
}
