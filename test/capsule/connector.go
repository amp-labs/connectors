package smartlead

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/capsule"
	"github.com/amp-labs/connectors/test/utils"
)

func GetCapsuleConnector(ctx context.Context) *capsule.Connector {
	filePath := credscanning.LoadPath(providers.Capsule)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	clientBuilder := &paramsbuilder.Client{}
	clientBuilder.WithApiKeyHeaderClient(ctx,
		http.DefaultClient, providers.Capsule,
		reader.Get(credscanning.Fields.ApiKey),
	)

	conn, err := capsule.NewConnector(
		parameters.Connector{
			AuthenticatedClient: clientBuilder.AuthClient.Caller.Client,
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
