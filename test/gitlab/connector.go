package gitlab

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gitlab"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *gitlab.Connector {
	filePath := credscanning.LoadPath(providers.GitLab)

	reader := utils.MustCreateProvCredJSON(filePath, false, credscanning.Fields.Token)

	token := reader.Get(credscanning.Fields.Token)

	client, err := common.NewCustomAuthHTTPClient(ctx, common.WithCustomHeaders(common.Header{Key: "PRIVATE-TOKEN", Value: token}))
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := gitlab.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
