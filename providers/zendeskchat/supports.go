package zendeskchat

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func responseField(objectName string) string {
	switch objectName {
	case "chats", "incremental/chats":
		return "chats"
	case "incremental/agent_events":
		return "agent_events"
	case "incremental/agent_timeline":
		return "agent_timeline"
	case "incremental/conversions":
		return "conversions"
	case "incremental/department_events":
		return "department_events"
	default:
		return ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"account", "agents", "bans", "bans/ip", "chats", "departments", "goals",
		"incremental/agent_events", "incremental/agent_timeline",
		"incremental/chats", "incremental/conversions", "incremental/department_events",
		"roles", "routing_settings/agents", "shortcuts", "skills", "triggers",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
