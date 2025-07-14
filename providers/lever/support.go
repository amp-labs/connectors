package lever

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput { // nolint:funlen
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
		"form_templates",
		"requisitions",
		"requisition_fields",
		"uploads",
		"users",
		"feedback_templates",
		"contacts",
		"opportunities",
		"postings",
	}

	deleteSupport := []string{
		"feedback_templates",
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
