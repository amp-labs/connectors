package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// Grouping of ObjectName to response field name mappings by Module.
type fieldMappings map[common.ModuleID]handy.DefaultMap[string, string]

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = fieldMappings{ //nolint:gochecknoglobals
	ModuleTicketing: handy.NewDefaultMap(map[string]string{
		"ticket_audits":        "audits",
		"search":               "results", // This is "/api/v2/search"
		"satisfaction_reasons": "reasons",
	},
		func(key string) string {
			// Other response fields are named after Object.
			return key
		},
	),
	ModuleHelpCenter: handy.NewDefaultMap(map[string]string{
		"articles":        "results",
		"article_labels":  "labels",
		"community_posts": "results",
	}, func(key string) string {
		return key
	}),
}
