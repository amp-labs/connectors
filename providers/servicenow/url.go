package servicenow

import (
	"strings"

	"github.com/amp-labs/connectors/common"
)

const restAPIPrefix = "api"

func (c *Connector) buildURL(config common.ReadParams) (string, error) {
	if len(config.NextPage) > 0 {
		return config.NextPage.String(), nil
	}

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return "", err
	}

	url.WithQueryParam("sysparm_fields", strings.Join(config.Fields.List(), ","))

	return url.String(), nil
}
