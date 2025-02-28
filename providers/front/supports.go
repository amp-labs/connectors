package front

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"accounts", "channels", "contacts", "contact_lists",
		"conversations", "events", "inboxes", "knowledge_bases",
		"links", "rules", "message_templates", "message_template_folders",
		"shifts", "tags", "teammates", "teammate_groups", "teams",
	}

	writeSupports := []string{
		"accounts", "contacts", "contact_lists", "conversations",
		"inboxes", // Create an inbox in the default team (workspace)
		"channels", "comments", "signatures",
		"knowledge_bases", "links", "message_templates", "message_template_folders",
		"shifts", "tags", "teammate_groups", "teams",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupports, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
