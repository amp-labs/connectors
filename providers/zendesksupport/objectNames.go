package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

const (
	objectNameTickets  = "tickets"
	objectNameRequests = "requests"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var objectsUnsupportedWrite = map[common.ModuleID]datautils.Set[string]{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		"attribute_values",
		"instance_values",
		"ticket_events",
		"ticket_metric_events",
	),
}

var writeURLExceptions = map[common.ModuleID]datautils.Map[string, string]{ //nolint:gochecknoglobals
	common.ModuleRoot: {
		"attributes":    "/routing/attributes",
		"organizations": "/organizations",
		"tickets":       "/tickets",
		"users":         "/users",
	},
}

var objectsWithCustomFields = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewStringSet(
		// https://developer.zendesk.com/api-reference/ticketing/tickets/tickets/#json-format
		objectNameTickets,
		// https://developer.zendesk.com/api-reference/ticketing/tickets/ticket-requests/#json-format
		objectNameRequests,
	),
}
