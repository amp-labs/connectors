package auphan

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// reportObjects are read by running reports/execute with the object name as the report type.
// To add a report-backed object, add its Auphan report type here.
var reportObjects = datautils.NewSet( //nolint:gochecknoglobals
	"payments",
	"sales",
)

// listObjects are read from a paginated list endpoint.
var listObjects = map[string]string{ //nolint:gochecknoglobals
	"stores": "settings/stores.list",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := reportObjects.List()

	for objectName := range listObjects {
		readSupport = append(readSupport, objectName)
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
