package workday

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/workday"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldTenantName = credscanning.Field{ //nolint:gochecknoglobals
	Name:      "tenantName",
	PathJSON:  "metadata.tenantName",
	SuffixENV: "TENANT_NAME",
}

func GetWorkdayConnector(ctx context.Context) *workday.Connector {
	filePath := credscanning.LoadPath(providers.Workday)

	reader := utils.MustCreateProvCredJSON(filePath, true, fieldTenantName)

	conn, err := workday.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		Workspace:           reader.Get(credscanning.Fields.Workspace),
		Metadata: map[string]string{
			"tenantName": reader.Get(fieldTenantName),
		},
	})
	if err != nil {
		utils.Fail("error while creating workday connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://impl.workday.com/ccx/oauth2/authorize",
			TokenURL:  "https://impl.workday.com/ccx/oauth2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}
