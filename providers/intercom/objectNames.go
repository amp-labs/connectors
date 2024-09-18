package intercom

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/intercom/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSet( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"admins":        "admins",
	"teams":         "teams",
	"ticket_types":  "ticket_types",
	"events":        "events",
	"segments":      "segments",
	"activity_logs": "activity_logs",
},
	func(key string) string {
		// Other objects are mapped to `data`.
		return "data"
	},
)
