package awsic

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/awsic"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetAWSIdentityCenterConnector(ctx context.Context) *awsic.Connector {
	filePath := credscanning.LoadPath(providers.AWSIdentityCenter)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	awsRegion := "us-east-2"

	client, err := common.NewAWSClient(ctx,
		http.DefaultClient,
		reader.Get(credscanning.Fields.Username),
		reader.Get(credscanning.Fields.Password),
		awsRegion,
	)

	if err != nil {
		testUtils.Fail(err.Error())
	}

	conn, err := awsic.NewConnector(
		common.Parameters{
			AuthenticatedClient: client,
			Metadata: map[string]string{
				"region": awsRegion,
				// TODO is it right to have ID on the connector level?
				// How many stores people have per region?
				"IdentityStoreID": "d-9a670e6550",
				"InstanceArn":     "arn:aws:sso:::instance/ssoins-668432c2a02ced8f",
			},
		},
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}
