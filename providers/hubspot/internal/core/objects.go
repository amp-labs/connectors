package core

import "github.com/amp-labs/connectors/internal/datautils"

const (
	// DefaultPageSize is the default page size for paginated requests.
	// HubSpot's read endpoints support max 100 records per page.
	DefaultPageSize    = "100"
	DefaultPageSizeInt = int64(100)
)

//nolint:gochecknoglobals
var (

	// ObjectsWithoutPropertiesAPISupport is a list of objectsNames that are part of Hubspot CRM API namespace,
	// but exist outside Hubspot Object Properties APIs.
	//
	// These objects cannot be accessed via `~/crm/v3/objects/{objectTypeId}/...`
	//
	// On the contrary those objects that are part of Object Properties APIs can be found
	// in the table here https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
	ObjectsWithoutPropertiesAPISupport = datautils.NewSet( //nolint:gochecknoglobals
		"lists",
	)

	// ObjectsMarketingAPI is a list of objectNames that are part of Hubspot Marketing API.
	ObjectsMarketingAPI = datautils.NewSet(
		"campaigns",
	)
)
