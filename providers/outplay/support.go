package outplay

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

var objectAPIPath = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"prospect":        "prospect/search",
	"prospectaccount": "prospectaccount/search",
	"sequence":        "sequence/search",
	"call":            "call/search",
	"task":            "task/list",
	"callanalysis":    "callanalysis/list",
}, func(objectName string) string {
	return objectName
})

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"prospect",
		"prospectaccount",
		"sequence",
		"sequenceprospect",
		"call",
		"task",
		"callanalysis",
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
