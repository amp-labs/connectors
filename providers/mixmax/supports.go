package mixmax

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func responseField(objectName string) string {
	switch objectName {
	case "appointmentlinks/me", "userpreferences/me", "users/me":
		return "" // indicates we're reading data fields from the root level.
	default:
		return "results"
	}
}

func supportedOperations() components.EndpointRegistryInput {
	// The objecst salesforce/* are not supported as they have not been tested.
	readSupport := []string{
		"appointmentlinks/me", "userpreferences/me", "users/me",
		"codesnippets", "filerequests", "insightsreports", "integrations/commands",
		"integrations/enhancements", "integrations/linkresolvers", "integrations/sidebars",
		"livefeed", // livefeed has another field object with stats field
		"meetinginvites", "meetingtypes", "messages", "polls", "qa", "rules", "sequences",
		"sequencefolders", "snippets", "snippettags", "teams", "unsubscribes", "yesno",
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
