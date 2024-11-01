package zendesksupport

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"ticket_audits":        "audits",
	"search":               "results", // This is "/api/v2/search"
	"satisfaction_reasons": "reasons",
},
	func(key string) string {
		// Other response fields are named after Object.
		return key
	},
)
