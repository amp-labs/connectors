package lemlist

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func responseSchema(objectName string) (string, string) {
	// team --> an object
	// api/team/senders --> an array of objects
	// api/team/credits --> object
	// api/campaigns --> object campaigns array
	// api/activities --> an array of objects
	// api/unsubscirbes --> an array of objects
	// api/hooks --> an array of objects
	// api/database/filters -->array of objects
	// api/schema/people --> an object
	// api/schema/companies --> an object
	switch objectName {
	case "campaigns", "schedules":
		return object, objectName
	case "team/senders", "activities", "unsubscirbes", "hooks", "database/filters":
		return list, ""
	default:
		return object, ""
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"team", "team/senders", "team/credits", "campaigns",
		"activities", "unsubscribes", "hooks", "database/filters",
		"schema/people", "schema/companies", "schedules",
	}

	writeSupport := []string{
		"campaigns", "schedules", "unsubscribes/*", "hooks", "database/people",
		"database/companies", "tasks", "tasks/ignore",
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
