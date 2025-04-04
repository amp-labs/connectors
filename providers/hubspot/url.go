package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(path ...string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL(path...)
}

func (c *Connector) getCRMObjectsReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// NB: The final slash is just to emulate prior behavior in earlier versions
	// of this code. If it turns out to be unnecessary, remove it.
	url, err := c.getURL("objects", config.ObjectName+"/")
	if err != nil {
		return nil, err
	}

	makeCRMObjectsQueryValues(config, url)

	return url, nil
}

func (c *Connector) getCRMObjectsSearchURL(config SearchParams) (*urlbuilder.URL, error) {
	return c.getURL("objects", config.ObjectName, "search")
}

func (c *Connector) getCRMSearchURL(config searchCRMParams) (*urlbuilder.URL, error) {
	return c.getURL(config.ObjectName, "search")
}
