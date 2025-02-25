package servicenow

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"incident", "cmdb_ci_email_server", "cmdb_data_classification"} //nolint:lll

	return components.EndpointRegistryInput{
		ModuleTable: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
