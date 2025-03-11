package smartlead

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/capsule"
	"github.com/amp-labs/connectors/test/utils"
)

func GetCapsuleConnector(ctx context.Context) *capsule.Connector {
	filePath := credscanning.LoadPath(providers.Capsule)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client := http.DefaultClient
	client.Transport = &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		ForceAttemptHTTP2: false,
	}

	clientBuilder := &paramsbuilder.Client{}
	clientBuilder.WithApiKeyHeaderClient(ctx,
		client, providers.Capsule,
		reader.Get(credscanning.Fields.ApiKey),
	)

	conn, err := capsule.NewConnector(
		common.Parameters{
			AuthenticatedClient: clientBuilder.AuthClient.Caller.Client,
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}
