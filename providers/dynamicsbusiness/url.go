package dynamicsbusiness

import (
	"fmt"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Microsoft uses special symbology when making queries.
var queryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%40": "@",
	"%24": "$",
	"%2C": ",",
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	url.AddEncodingExceptions(queryEncodingExceptions)

	return url, nil
}

func (c *Connector) getMetadataURL() (*urlbuilder.URL, error) {
	return constructURL(c.ProviderInfo().BaseURL,
		"v2.0", c.tenantID,
		"/Production/api/v2.0/entityDefinitions")
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	return constructURL(c.ProviderInfo().BaseURL,
		"v2.0", c.tenantID,
		fmt.Sprintf("/Production/api/v2.0/companies(%v)", c.companyID),
		objectName)
}
