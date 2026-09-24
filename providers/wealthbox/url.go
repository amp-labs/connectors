package wealthbox

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const apiVersion = "v1"

// https://dev.wealthbox.com/#topics-custom-fields
func (c *Connector) getCustomFieldsURL(documentType string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "categories", "custom_fields")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("document_type", documentType)

	return url, nil
}
