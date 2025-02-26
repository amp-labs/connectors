package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var objectsUnsupportedWrite = map[common.ModuleID]datautils.Set[string]{ //nolint:gochecknoglobals
	ModuleTicketing: datautils.NewSet(
		"attribute_values",
		"instance_values",
		"ticket_events",
		"ticket_metric_events",
	),
	ModuleHelpCenter: datautils.NewStringSet(),
}

var writeURLExceptions = map[common.ModuleID]datautils.Map[string, string]{ //nolint:gochecknoglobals
	ModuleTicketing: {
		"attributes":    "/api/v2/routing/attributes",
		"organizations": "/api/v2/organizations",
		"tickets":       "/api/v2/tickets",
		"users":         "/api/v2/users",
	},
	ModuleHelpCenter: {},
}
