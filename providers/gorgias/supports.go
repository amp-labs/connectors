package gorgias

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"account", "customers", "custom-fields", "events", "integrations",
		"jobs", "macros", "rules", "satisfaction-surveys", "tags", "teams",
		"tickets", "messages", "users", "views", "phone/voice-calls",
		"phone/voice-call-recordings", "phone/voice-call-events",
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
