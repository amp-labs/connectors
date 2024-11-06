package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = common.ModuleObjectNameToFieldName{ //nolint:gochecknoglobals
	ModuleTicketing: datautils.NewDefaultMap(map[string]string{
		"ticket_audits":        "audits",
		"search":               "results", // This is "/api/v2/search"
		"satisfaction_reasons": "reasons",
	},
		func(objectName string) (fieldName string) {
			return objectName
		},
	),
	ModuleHelpCenter: datautils.NewDefaultMap(map[string]string{
		"articles":        "results",
		"article_labels":  "labels",
		"community_posts": "results",
	}, func(objectName string) (fieldName string) {
		return objectName
	}),
}
