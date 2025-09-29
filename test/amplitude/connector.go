package amplitude

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/amplitude"
	"github.com/amp-labs/connectors/test/utils"
)

func GetAmplitudeConnector(ctx context.Context) *amplitude.Connector {
	filePath := credscanning.LoadPath(providers.Amplitude)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, 
		reader.Get(credscanning.Fields.Username),
		reader.Get(credscanning.Fields.Password),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := amplitude.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("create amplitude connector", "error: ", err)
	}

	return conn
}
