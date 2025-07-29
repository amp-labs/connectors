package braze

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}
	writeSupport := []string{"*"}

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

var readEndpointsByObject = map[string]string{ //nolint: gochecknoglobals
	"campaigns":           "campaigns/list",
	"canvas":              "canvas/list",
	"segments":            "segments/list",
	"preference_center":   "preference_center/v1/list",
	"subscription/status": "subscription/status/get",
	"content_blocks":      "content_blocks/list",
	"templates/email":     "templates/email/list",
}

var writeEndpointsByObject = map[string]string{ //nolint: gochecknoglobals
	"subscription/status": "subscription/status/set",
	"preference_center":   "preference_center/v1",
	"content_blocks":      "content_blocks/create",
	"templates/email":     "templates/email/create",
}
