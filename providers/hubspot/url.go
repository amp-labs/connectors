package hubspot

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

// This file defines HubSpot endpoint builders used by this connector.
//
// HubSpot exposes several API families. Many connector operations live under
// the CRM namespace, but not all CRM endpoints belong to the HubSpot Objects API.
//
// In Ampersand terminology, "object" refers to a connector resource such as contacts or lists.
// In this file, "Objects" refers specifically to the HubSpot Objects API.
// It does not refer to an Ampersand object.
//
//                 +-----------------------------------+
//                 |               CRM                 |
//                 |   +---------------------------+   |
//                 |   |      HubSpot Objects      |   |
//                 |   |  Contacts                 |   |
//                 |   |  Leads                    |   |
//                 |   |  Quotes                   |   |
//                 |   +---------------------------+   |
//                 |  Lists                            |
//                 +-----------------------------------+
//
// Example:
//
//   - contacts belong to both the CRM namespace and the Objects API
//   - lists belong to the CRM namespace but not to the Objects API
//
// The distinction matters because URL layouts differ between these endpoint families.

// Used by GetPostAuthInfo.
// https://developers.hubspot.com/docs/api-reference/latest/account/account-information/get-account-details
func (c *Connector) getAccountDetailsURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "account-info", core.APIVersion2026March, "details")
}

// Returns the schema endpoint for an object definition.
//
// Used to construct object metadata.
// Output: schemaResponse.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/schemas/get-schema
func (c *Connector) getCRMSchemaURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "crm-object-schemas", core.APIVersion2026March, "schemas", objectName)
}

// Returns the properties endpoint for an object.
//
// Used to construct object field metadata.
// Output: fieldDescriptionResponse.
//
// This endpoint does not currently expose APIVersion2026March, so it still uses v3.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/get-properties
func (c *Connector) getCRMPropertiesURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "properties", objectName, "/")
}

// Returns the base HubSpot Objects API endpoint for CRUD operations.
//
// This URL shape is shared by CRUD operations.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/get-contacts
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/create-contact
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/update-contact
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/delete-contact
//
// NOTE: the path layout still follows the older v3 structure.
func (c *Connector) getCRMObjectsURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName)
}

// Returns the delete endpoint for the HubSpot Objects API.
//
// TODO: replace this helper with getCRMObjectsURL once getCRMObjectsURL is migrated to APIVersion2026March.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/delete-contact
func (c *Connector) getCRMObjectsDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return c.crmURL("objects", core.APIVersion2026March, objectName, recordID)
}

// Returns the batch read endpoint for the HubSpot Objects API.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/batch/get-contacts
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/leads/batch/get-leads
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/quotes/batch/get-quotes
func (c *Connector) getCRMObjectsBatchReadURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName, "batch", "read")
}

// Returns the search endpoint for the HubSpot Objects API.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/search/search-contacts
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/leads/search/search-leads
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/quotes/search/search-quotes
func (c *Connector) getCRMObjectsSearchURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName, "search")
}

// Returns a CRM search endpoint that does not belong to the Objects API.
//
// Some CRM resources, such as lists, live under CRM but outside the
// Objects API and therefore follow a different URL layout.
//
// https://developers.hubspot.com/docs/api-reference/latest/crm/lists/guide#retrieve-by-searching-list-details
func (c *Connector) getCRMSearchURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, objectName, "search")
}

// Returns the base HubSpot Marketing API endpoint for CRUD operations.
//
// This URL shape is shared by CRUD operations.
//
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/get-campaigns
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/create-campaign
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/update-campaign
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/delete-campaign
func (c *Connector) getMarketingURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "marketing", objectName, core.APIVersion2026March)
}

func (c *Connector) crmURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{"crm"}, paths...)

	// URL: "https://api.hubapi.com/crm"
	return urlbuilder.New(c.ProviderInfo().BaseURL, parts...)
}

func (c *Connector) getURLFromRoot(relativePath string) string {
	return c.ProviderInfo().BaseURL + relativePath
}
