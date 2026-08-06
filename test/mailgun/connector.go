package mailgun

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

// fieldRegion selects the Mailgun API region ("eu" for the EU base URL). It is
// optional; the US base URL is used when unset.
var fieldRegion = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "region",
	PathJSON:  "metadata.region",
	SuffixENV: "REGION",
}

// GetWorkspace returns the Mailgun sending domain from the credentials file.
func GetWorkspace() string {
	filePath := credscanning.LoadPath(providers.Mailgun)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldRegion)

	return reader.Get(credscanning.Fields.Workspace)
}

func GetMailgunConnector(ctx context.Context) *mailgun.Connector {
	filePath := credscanning.LoadPath(providers.Mailgun)
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldRegion)

	metadata := make(map[string]string)
	if region := reader.Get(fieldRegion); region != "" {
		metadata["region"] = region
	}

	// Mailgun uses HTTP Basic Auth ("api" as username, API key as password).
	client := utils.NewBasicAuthClient(ctx, reader)

	conn, err := mailgun.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		// Workspace is the Mailgun sending domain, required for domain-scoped
		// objects (bounces, complaints, templates, ...).
		Workspace: reader.Get(credscanning.Fields.Workspace),
		Metadata:  metadata,
	})
	if err != nil {
		utils.Fail("error creating mailgun connector", "error", err)
	}

	return conn
}
