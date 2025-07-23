package zendeskchat

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
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

var updatesByPatch = map[string]string{ //nolint:gochecknoglobals
	"routing_settings/account":   "routing_settings/account",
	"routing_settings/agents":    "routing_settings/agents",
	"routing_settings/agents/me": "routing_settings/agents/me",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"account", "agents", "bans", "bans/ip", "chats", "departments", "goals",
		"incremental/agent_events", "incremental/agent_timeline",
		"incremental/chats", "incremental/conversions", "incremental/department_events",
		"roles", "routing_settings/agents", "shortcuts", "skills", "triggers",
		"oauth/clients", "oauth/tokens",
	}

	writeSupport := []string{
		"account", "bans", "visitors", "triggers", "oauth/clients",
		"shortcuts", "goals", "skills", "roles", "chats",
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
		},
	}
}
