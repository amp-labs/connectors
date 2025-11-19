package loxo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/loxo"
	"github.com/amp-labs/connectors/test/utils"
)

var fieldWorkspace = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "workspace",
	PathJSON:  "metadata.workspace",
	SuffixENV: "WORKSPACE",
}

var fieldAgencySlug = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "agencySlug",
	PathJSON:  "metadata.agencySlug",
	SuffixENV: "AGENCY_SLUG",
}

func GetLoxoConnector(ctx context.Context) *loxo.Connector {
	filePath := credscanning.LoadPath(providers.Loxo)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldWorkspace, fieldAgencySlug)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Loxo)

	conn, err := loxo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           reader.Get(fieldWorkspace),
		Metadata: map[string]string{
			"agencySlug": reader.Get(fieldAgencySlug),
		},
	})
	if err != nil {
		utils.Fail("error creating Loxo App connector", "error", err)
	}

	return conn
}
