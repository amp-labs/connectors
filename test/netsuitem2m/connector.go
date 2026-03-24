package netsuitem2m

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	netsuitem2m "github.com/amp-labs/connectors/providers/netsuite/m2m"
	"github.com/amp-labs/connectors/test/utils"
)

// Custom credential fields for M2M auth (not in the standard Fields struct).
var (
	fieldCertificateID = credscanning.Field{
		Name:      "certificateId",
		PathJSON:  "certificateId",
		SuffixENV: "CERTIFICATE_ID",
	}
	fieldPrivateKey = credscanning.Field{
		Name:      "privateKey",
		PathJSON:  "privateKey",
		SuffixENV: "PRIVATE_KEY",
	}
)

func GetNetsuiteM2MRESTAPIConnector(ctx context.Context) *netsuitem2m.Connector {
	return getConnector(ctx, providers.ModuleNetsuiteRESTAPI)
}

func GetNetsuiteM2MSuiteQLConnector(ctx context.Context) *netsuitem2m.Connector {
	return getConnector(ctx, providers.ModuleNetsuiteSuiteQL)
}

func GetNetsuiteM2MRESTletConnector(ctx context.Context) *netsuitem2m.Connector {
	return getConnector(ctx, providers.ModuleNetsuiteRESTlet)
}

func getConnector(ctx context.Context, module common.ModuleID) *netsuitem2m.Connector {
	reader := getJSONReader()

	accountID := reader.Get(credscanning.Fields.Workspace)
	clientID := reader.Get(credscanning.Fields.ClientId)
	certificateID := reader.Get(fieldCertificateID)
	privateKeyPEM := reader.Get(fieldPrivateKey)

	client, err := netsuitem2m.NewM2MAuthenticatedClient(ctx, accountID, clientID, certificateID, privateKeyPEM)
	if err != nil {
		utils.Fail("error creating M2M auth client", "error", err)
	}

	conn, err := netsuitem2m.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           accountID,
		Module:              module,
	})
	if err != nil {
		utils.Fail("error creating netsuite M2M connector", "error", err)
	}

	return conn
}

func getJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.NetsuiteM2M)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldCertificateID, fieldPrivateKey)

	return reader
}
