package pipeliner

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
)

var fieldRegion = credscanning.Field{
	Name:      "region",
	PathJSON:  "metadata.region",
	SuffixENV: "REGION",
}

func GetPipelinerConnector(ctx context.Context) *pipeliner.Connector {
	filePath := credscanning.LoadPath(providers.Pipeliner)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldRegion)

	metadata := make(map[string]string)
	if region := reader.Get(fieldRegion); region != "" {
		metadata["region"] = region
	}

	conn, err := pipeliner.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
			Metadata:            metadata,
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
