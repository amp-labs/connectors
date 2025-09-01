package breakcold

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"status",
		"workspaces",
		"members",
		"leads",
		"tags",
		"lists",
		"notes",
		"reminders",
	}

	writeSupport := []string{
		"status",
		"lead",
		"leads/add-list",
		"tags",
		"lists",
		"notes",
		"reminders",
		"attribute",
	}

	deleteSupport := []string{
		"status",
		"lead",
		"tags",
		"lists",
		"notes",
		"reminders",
		"attribute",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
