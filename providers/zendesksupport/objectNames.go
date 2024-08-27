package zendesksupport

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSet( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = handy.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"ticket_audits":        "audits",
	"search":               "results", // This is "/api/v2/search"
	"satisfaction_reasons": "reasons",
},
	func(key string) string {
		// Other response fields are named after Object.
		return key
	},
)
