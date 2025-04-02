package heyreach

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/test/utils"
)

func GetHeyreachConnector(ctx context.Context) *heyreach.Connector {
	filePath := credscanning.LoadPath(providers.HeyReach)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := heyreach.NewConnector(
		heyreach.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating heyreach connector", "error", err)
	}

	return conn
}
