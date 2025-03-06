package capsule

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
	"github.com/spyzhov/ajson"
)

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {{
			Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
			Support:  components.ReadSupport,
		}},
	}
}

func getNextRecordsURL(_ *ajson.Node) (string, error) {
	// Pagination is not supported for this provider.
	return "", nil
}
