package drift

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	list   = "list"
	object = "object"
	data   = "data"
)

func responseSchema(objectName string) (string, string) {
	switch objectName {
	case "users/list", "conversations/list", "teams/org", "users/meetings/org":
		return object, data
	case "playbooks/list", "playbooks/clp":
		return list, ""
	default:
		return object, ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"users/list", "conversations/list", "teams/org", "users/meetings/org",
		"playbooks/list", "playbooks/clp", "conversations/stats", "scim/Users",
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
