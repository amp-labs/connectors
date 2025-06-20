package lever

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"feedback",
		"files",
		"interviews",
		"notes",
		"offers",
		"panels",
		"forms",
		"referrals",
		"resumes",
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

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
