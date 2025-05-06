package hubspot

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) getCRMObjectsReadURL(params common.ReadParams) (string, error) {
	// NB: The final slash is just to emulate prior behavior in earlier versions
	// of this code. If it turns out to be unnecessary, remove it.
	url, err := c.ModuleAPI.URL("objects", params.ObjectName+"/")
	if err != nil {
		return "", err
	}

	fields := params.Fields.List()
	if len(fields) != 0 {
		url.WithQueryParam("properties", strings.Join(fields, ","))
	}

	if params.Deleted {
		url.WithQueryParam("archived", "true")
	}

	url.WithQueryParam("limit", DefaultPageSize)

	return url.String(), nil
}

func (c *Connector) getCRMObjectsSearchURL(config SearchParams) (*urlbuilder.URL, error) {
	return c.ModuleAPI.URL("objects", config.ObjectName, "search")
}

func (c *Connector) getCRMSearchURL(config searchCRMParams) (*urlbuilder.URL, error) {
	return c.ModuleAPI.URL(config.ObjectName, "search")
}
