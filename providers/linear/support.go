package linear

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"attachments",
		"auditEntries",
		"comments",
		"customers",
		"cycles",
		"documents",
		"favorites",
		"initiatives",
		"issues",
		"notifications",
		"projects",
		"projectStatuses",
		"teamMemberships",
		"teams",
		"triageResponsibilities",
		"users",
		"workflowStates",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
