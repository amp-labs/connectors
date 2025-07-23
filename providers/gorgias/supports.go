package gorgias

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"account", "customers", "custom-fields", "events", "integrations",
		"jobs", "macros", "rules", "satisfaction-surveys", "tags", "teams",
		"tickets", "messages", "users", "views", "phone/voice-calls",
		"phone/voice-call-recordings", "phone/voice-call-events", "widgets",
	}

	writeSupport := []string{
		"account/settings", "customers", "custom-fields", "integrations", "jobs",
		"macros", "rules", "satisfaction-surveys",
		"search", // https://developers.gorgias.com/reference/search-1
		"tags", "teams", "tickets", "users", "views", "widgets",
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
