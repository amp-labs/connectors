package lever

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"archive_reasons",
		"audit_events",
		"sources",
		"stages",
		"tags",
		"users",
		"feedback_templates",
		"opportunities",
		"postings",
		"form_templates",
		"requisitions",
		"requisition_fields",
	}

	writeSupport := []string{
		"notes",
		"addLinks",
		"removeLinks",
		"addTags",
		"removeTags",
		"addSources",
		"removeSources",
		"forms",
		"form_templates",
		"requisitions",
		"requisition_fields",
		"uploads",
		"users",
		"feedback_templates",
		"contacts",
		"stage",
		"archived",
	}

	deleteSupport := []string{
		"feedback_templates",
		"notes",
		"form_templates",
		"requisitions",
		"requisition_fields",
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
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
