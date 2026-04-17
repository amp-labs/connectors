package odoo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/odoo"
	"github.com/amp-labs/connectors/test/utils"
)

var fieldOdooDomain = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "odoo_domain",
	PathJSON:  "metadata.odoo_domain",
	SuffixENV: "ODOO_DOMAIN",
}

// GetConnector builds an Odoo connector from ./odoo-creds.json or ODOO_CRED_FILE.
func GetConnector(ctx context.Context) *odoo.Connector {
	filePath := credscanning.LoadPath(providers.Odoo)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldOdooDomain)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Odoo)

	domain := reader.Get(fieldOdooDomain)
	if domain == "" {
		utils.Fail("missing metadata.odoo_domain in creds")
	}

	conn, err := odoo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"odoo_domain": domain,
		},
	})
	if err != nil {
		utils.Fail("error creating Odoo connector", "error", err)
	}

	return conn
}
