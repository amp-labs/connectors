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
	reader := getAmplitudeJSONReader()

	conn, err := amplitude.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
	})
	if err != nil {
		utils.Fail("create amplitude connector", "error: ", err)
	}

	return conn
}

func GetAmplitudeAPIkey() common.AuthToken {
	reader := getAmplitudeJSONReader()

	return common.AuthToken(reader.Get(credscanning.Fields.Username))
}

func getAmplitudeJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Amplitude)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	return reader
}
