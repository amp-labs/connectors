package hubspot

import "github.com/amp-labs/connectors/internal/datautils"

// This is a list of objectsNames that are part of Hubspot CRM Module but exist outside Object Properties APIs.
//
// These objects cannot be accessed via `~/crm/v3/objects/{objectTypeId}/...`
//
// On the contrary those objects that are part of Object Properties APIs can be found
// in the table here https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
var crmObjectsWithoutPropertiesAPISupport = datautils.NewSet( //nolint:gochecknoglobals
	"lists",
)
