package hubspot

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

func (c *Connector) getBatchReadURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName, "batch", "read")
}

func (c *Connector) getCRMObjectsReadURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName)
}

func (c *Connector) getCRMObjectsSearchURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "objects", objectName, "search")
}

func (c *Connector) getCRMSearchURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, objectName, "search")
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/properties/get-properties
// Note: Version APIVersion2026March is NOT FOUND at the moment for this endpoint. Using older V3.
func (c *Connector) getPropertiesURL(objectName string) (*urlbuilder.URL, error) {
	return c.crmURL(core.APIVersion3, "properties", objectName, "/")
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/schemas/get-schema
func (c *Connector) getObjectSchemaURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "crm-object-schemas", core.APIVersion2026March, "schemas", objectName)
}

// https://developers.hubspot.com/docs/api-reference/latest/account/account-information/get-account-details
func (c *Connector) getAccountDetailsURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, "account-info", core.APIVersion2026March, "details")
}

func (c *Connector) getURLFromRoot(relativePath string) string {
	return c.ProviderInfo().BaseURL + relativePath
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/delete-contact
func (c *Connector) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return c.crmURL("objects", core.APIVersion2026March, objectName, recordID)
}

func (c *Connector) crmURL(paths ...string) (*urlbuilder.URL, error) {
	parts := append([]string{"crm"}, paths...)

	// URL: "https://api.hubapi.com/crm"
	return urlbuilder.New(c.ProviderInfo().BaseURL, parts...)
}
