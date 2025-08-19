package drift

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	list          = "list"
	object        = "object"
	data          = "data"
	updateAccount = "accounts/update"
)

func responseSchema(objectName string) (string, string) {
	switch objectName {
	case "users", "conversations", "teams/org", "users/meetings/org":
		return object, data
	case "playbooks", "playbooks/clp":
		return list, ""
	default:
		return object, ""
	}
}

func writeResponseField(objectName string) string {
	switch objectName {
	case "contacts":
		return "data"
	default:
		return ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"users", "conversations", "teams/org", "users/meetings/org",
		"playbooks", "playbooks/clp", "conversations/stats", "scim/Users",
	}

	writeSupport := []string{
		"contacts", "emails/unsubscribe", "contacts/timeline", "conversations/new", "accounts/create",
		"accounts/update", // updates do not need recordIdPath
		"scim/Users",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			}, {
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
