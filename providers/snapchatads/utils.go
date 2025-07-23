package snapchatads

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "organizations", c.organizationId, objName)
	if err != nil {
		return nil, err
	}

	return url, nil
}
