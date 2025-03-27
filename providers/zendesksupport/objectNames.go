package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var objectsUnsupportedWrite = map[common.ModuleID]datautils.Set[string]{ //nolint:gochecknoglobals
	common.ModuleID(providers.ModuleZendeskTicketing): datautils.NewSet(
		"attribute_values",
		"instance_values",
		"ticket_events",
		"ticket_metric_events",
	),
	common.ModuleID(providers.ModuleZendeskHelpCenter): datautils.NewStringSet(),
}

var writeURLExceptions = map[common.ModuleID]datautils.Map[string, string]{ //nolint:gochecknoglobals
	common.ModuleID(providers.ModuleZendeskTicketing): {
		"attributes":    "/api/v2/routing/attributes",
		"organizations": "/api/v2/organizations",
		"tickets":       "/api/v2/tickets",
		"users":         "/api/v2/users",
	},
	common.ModuleID(providers.ModuleZendeskHelpCenter): {},
}
