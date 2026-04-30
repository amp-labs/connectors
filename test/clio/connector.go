package clio

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	clioConnector "github.com/amp-labs/connectors/providers/clio"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var clioRegionField = credscanning.Field{ // nolint:gochecknoglobals
	Name:      "region",
	PathJSON:  "metadata.region",
	SuffixENV: "REGION",
}

func GetClioManageConnector(ctx context.Context) *clioConnector.Connector {
	return getClioConnector(ctx, providers.ModuleClioManage)
}

func GetClioGrowConnector(ctx context.Context) *clioConnector.Connector {
	return getClioConnector(ctx, providers.ModuleClioGrow)
}

func getClioConnector(ctx context.Context, moduleID common.ModuleID) *clioConnector.Connector {
	filePath := credscanning.LoadPath(providers.Clio, moduleID)
	reader := utils.MustCreateProvCredJSON(filePath, true, clioRegionField)

	workspace := reader.Get(credscanning.Fields.Workspace)
	region := reader.Get(clioRegionField)

	conn, err := clioConnector.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		Workspace:           workspace,
		Metadata: map[string]string{
			"region": region,
		},
		Module: moduleID,
	})
	if err != nil {
		utils.Fail("error creating clio connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)
	region := reader.Get(clioRegionField)
	regionPrefix := clioRegionPrefix(region)

	authHost := workspace
	if workspace == "api.clio.com" {
		authHost = "auth.api.clio.com"
	}

	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		// tests use existing credentials and token refresh;
		// they do not execute an OAuth authorization-code redirect flow.
		RedirectURL: fmt.Sprintf("https://%s%s", regionPrefix, authHost),
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s%s/oauth/authorize", regionPrefix, authHost),
			TokenURL:  fmt.Sprintf("https://%s%s/oauth/token", regionPrefix, authHost),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func clioRegionPrefix(region string) string {
	if region == "" || strings.EqualFold(region, "us") {
		return ""
	}

	return region + "."
}
