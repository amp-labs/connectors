package capsule

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/endpoints"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations(catalog *endpoints.Catalog) components.EndpointRegistryInput {
	readSupport := catalog.ReadOperation.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {{
			Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
			Support:  components.ReadSupport,
		}},
	}
}
