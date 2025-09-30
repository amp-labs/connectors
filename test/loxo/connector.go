package loxo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/loxo"
	"github.com/amp-labs/connectors/test/utils"
)

var fieldDomain = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "domain",
	PathJSON:  "substitutions.domain",
	SuffixENV: "DOMAIN",
}

var fieldAgencySlug = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "agency_slug",
	PathJSON:  "substitutions.agency_slug",
	SuffixENV: "AGENCY_SLUG",
}

func GetLoxoConnector(ctx context.Context) *loxo.Connector {
	filePath := credscanning.LoadPath(providers.Loxo)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldDomain, fieldAgencySlug)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Loxo)

	conn, err := loxo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"domain":      reader.Get(fieldDomain),
			"agency_slug": reader.Get(fieldAgencySlug),
		},
	})
	if err != nil {
		utils.Fail("error creating Loxo App connector", "error", err)
	}

	return conn
}
