package capsule

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/endpoints"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations(catalog *endpoints.Catalog) components.EndpointRegistryInput {
	readSupport := catalog.ReadOperation.ObjectNames().GetList(staticschema.RootModuleID)
	writeSupport := datautils.MergeUniqueLists(
		catalog.CreateOperation.ObjectNames(),
		catalog.UpdateOperation.ObjectNames(),
	).GetList(staticschema.RootModuleID)
	deleteSupport := catalog.DeleteOperation.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {{
			Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
			Support:  components.ReadSupport,
		}, {
			Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
			Support:  components.WriteSupport,
		}, {
			Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
			Support:  components.DeleteSupport,
		}},
	}
}
