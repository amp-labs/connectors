package phoneburner

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"contacts",
		"customfields",
		"dialsession",
		"folders",
		"members",
		"tags",
		"voicemails",
	}

	writeSupport := []string{
		"contacts",
		"customfields",
		"dialsession",
		"folders",
		"members",
		"tags",
	}

	deleteSupport := []string{
		"contacts",
		"customfields",
		"folders",
		"members",
		"phonenumber",
		"tags",
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
