package solarwinds

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/solarwinds"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSolarWindsConnector(ctx context.Context) *solarwinds.Connector {
	filePath := credscanning.LoadPath(providers.SolarWindsServiceDesk)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.SolarWindsServiceDesk)

	conn, err := solarwinds.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"subdomain": "api",
		},
	})
	if err != nil {
		utils.Fail("error creating Stripe connector", "error", err)
	}

	return conn
}
