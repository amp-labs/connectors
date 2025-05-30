package pipeliner

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
)

func GetPipelinerConnector(ctx context.Context) *pipeliner.Connector {
	filePath := credscanning.LoadPath(providers.Pipeliner)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := pipeliner.NewConnector(
		pipeliner.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.Username),
			reader.Get(credscanning.Fields.Password),
		),
		pipeliner.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
