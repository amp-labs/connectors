package auphan

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

// reportObjects are read by running reports/execute with the object name as the report type.
// The value is the report's "selection" (the date field it filters on), or "" for reports
// without a selection filter, which reject it with UNKNOWN_FILTER (petty_cash filters on tracking_date).
// To add a report-backed object, add its Auphan report type here.
var reportObjects = datautils.Map[string, string]{ //nolint:gochecknoglobals
	"payments":   "close_date",
	"sales":      "close_date",
	"petty_cash": "",
}

// listObjects are read from a paginated list endpoint.
var listObjects = map[string]string{ //nolint:gochecknoglobals
	"stores": "settings/stores.list",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := reportObjects.Keys()

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
