package shared

import "github.com/amp-labs/connectors/internal/datautils"

const (
	// DefaultPageSize is the default page size for paginated requests.
	// HubSpot's read endpoints support max 100 records per page.
	//
	// Reference for CRM module:
	// https://developers.hubspot.com/docs/api-reference/latest/crm/search-the-crm#limits
	//
	// Reference for Marketing module:
	// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/guide#search-for-campaigns
	DefaultPageSize    = "100"
	DefaultPageSizeInt = int64(100)
)

// ObjectsWithoutPropertiesAPISupport is a list of objectsNames that are part of Hubspot CRM Module
// but exist outside Object Properties APIs.
//
// These objects cannot be accessed via `~/crm/v3/objects/{objectTypeId}/...`
//
// On the contrary those objects that are part of Object Properties APIs can be found
// in the table here https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
var ObjectsWithoutPropertiesAPISupport = datautils.NewSet( //nolint:gochecknoglobals
	"lists",
)
