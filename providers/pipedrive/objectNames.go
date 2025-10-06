package pipedrive

import "github.com/amp-labs/connectors/internal/datautils"

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
	func(key string) string {
		return key
	},
)
