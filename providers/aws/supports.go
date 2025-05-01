package aws

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws/internal/identitystore"
	"github.com/amp-labs/connectors/providers/aws/internal/ssoadmin"
)

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		providers.ModuleAWSIdentityCenter: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(
					datautils.MergeSets(
						identitystore.ReadObjectCommands.KeySet(),
						ssoadmin.ReadObjectCommands.KeySet(),
					).List(), ",")),
				Support: components.ReadSupport,
			},
		},
	}
}
