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

	// CRMObjectsWithoutPropertiesAPISupport contains HubSpot CRM object names that
	// belong to the CRM API namespace but are not supported by the Object Properties API.
	//
	// These objects are not accessible through endpoints under either:
	//   /crm/v3/objects/{objectTypeId}/
	//	 /crm/objects/2026-03/{objectType}/{objectId}
	//
	// Objects that do support the Object Properties API are listed here:
	// https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
	CRMObjectsWithoutPropertiesAPISupport = datautils.NewSet( //nolint:gochecknoglobals
		"lists",
	)

	// MarketingObjects contains object names that belong to the HubSpot Marketing API.
	//
	// The Marketing API is separate from the CRM API and is not related to the Objects API.
	MarketingObjects = datautils.NewSet(
		"campaigns",
	)
)
