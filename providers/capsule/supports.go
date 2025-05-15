package capsule

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
)

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)
	writeSupport := datautils.MergeUniqueLists(
		supportedObjectsByCreate,
		supportedObjectsByUpdate,
	).GetList(common.ModuleRoot)
	deleteSupport := supportedObjectsByDelete[common.ModuleRoot].List()

	return components.EndpointRegistryInput{
		common.ModuleRoot: {{
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
