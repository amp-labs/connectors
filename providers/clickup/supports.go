package clickup

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/clickup/metadata"
)

func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
