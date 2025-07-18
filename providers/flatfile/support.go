package flatfile

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// nolint: gochecknoglobals
var supportObjectSince = datautils.NewSet(
	"events",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := schemas.ObjectNames().GetList(common.ModuleRoot)

	writeSupport := []string{
		"apps",
		"prompts",
		"environments",
		"events",
		"files",
		"jobs",
		"mapping",
		"spaces",
		"workbooks",
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
		},
	}
}
