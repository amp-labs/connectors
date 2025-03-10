package clari

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func responseField(objectName string) string {
	switch objectName {
	case "audit/events":
		return "items"
	case "export/jobs":
		return "jobs"
	default:
		return ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"export/jobs", "audit/events", "admin/limits",
	}

	writeSupport := []string{
		"export/jobs", "export/audit/events", "export/activity",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
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
