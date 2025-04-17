package groove

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func responseField(objectName string) string {
	switch objectName {
	case "kb":
		return "knowledge_bases"
	case "kb/themes":
		return "themes"
	case "tickets/count":
		return ""
	default:
		return objectName
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"tickets", "customers", "mailboxes", "folders",
		"agents", "groups", "kb/themes", "widgets",
		"kb", "tickets/count",
	}

	writeSupport := []string{
		"tickets", "webhooks", "groups", "widgets",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			}, {
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
