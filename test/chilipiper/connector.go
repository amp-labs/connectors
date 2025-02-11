package chilipiper

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/chilipiper"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetChiliPiperConnector(ctx context.Context) *chilipiper.Connector {
	filePath := credscanning.LoadPath(providers.ChiliPiper)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := chilipiper.NewConnector(
		chilipiper.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		testUtils.Fail("error creating Chilipiper connector", "error", err)
	}

	return conn
}
