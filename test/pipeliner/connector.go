package pipeliner

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	pipeliner2 "github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
)

func GetPipelinerConnector(ctx context.Context) *pipeliner2.Connector {
	filePath := credscanning.LoadPath(providers.Pipeliner)
	reader := utils.MustCreateProvCredJSON(filePath, false, true)

	conn, err := pipeliner2.NewConnector(
		pipeliner2.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.Username),
			reader.Get(credscanning.Fields.Password),
		),
		pipeliner2.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
