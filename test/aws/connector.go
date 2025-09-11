package aws

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

/*
File example:
	{
	  "provider": "aws",
	  "username": ".....", // AWS Access Key ID
	  "password": ".....", // AWS Access Key Secret
	  "metadata": {
		"region": "us-east-2",
		"identityStoreId": "d-9a670e6550",
		"instanceARN": "arn:aws:sso:::instance/ssoins-668432c2a02ced8f"
	  }
	}
*/

// nolint:gochecknoglobals
var (
	fieldRegion = credscanning.Field{
		Name:      "region",
		PathJSON:  "metadata.region",
		SuffixENV: "REGION",
	}
	fieldIdentityStoreID = credscanning.Field{
		Name:      "identityStoreId",
		PathJSON:  "metadata.identityStoreId",
		SuffixENV: "IDENTITY_STORE_ID",
	}
	fieldInstanceArn = credscanning.Field{
		Name:      "instanceARN",
		PathJSON:  "metadata.instanceARN",
		SuffixENV: "INSTANCE_ARN",
	}
)

func GetAWSConnector(ctx context.Context, module common.ModuleID) *aws.Connector {
	filePath := credscanning.LoadPath(providers.AWS)
	reader := testUtils.MustCreateProvCredJSON(filePath, false,
		fieldRegion, fieldIdentityStoreID, fieldInstanceArn,
	)

	awsRegion := reader.Get(fieldRegion)

	client, err := common.NewAWSClient(ctx,
		http.DefaultClient,
		reader.Get(credscanning.Fields.Username),
		reader.Get(credscanning.Fields.Password),
		awsRegion,
	)
	if err != nil {
		testUtils.Fail(err.Error())
	}

	conn, err := aws.NewConnector(
		common.ConnectorParams{
			Module:              module,
			AuthenticatedClient: client,
			Metadata: map[string]string{
				"region":          awsRegion,
				"identityStoreId": reader.Get(fieldIdentityStoreID),
				"instanceARN":     reader.Get(fieldInstanceArn),
			},
		},
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}
