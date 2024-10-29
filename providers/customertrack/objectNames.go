package customertrack

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/customerapp/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"object_types":        "types",
	"transactional":       "messages",
	"subscription_topics": "topics",
},
	func(key string) string {
		return key
	},
)
