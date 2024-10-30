package zohocrm

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// Check if we're reading the next-page.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	// Object names in ZohoCRM API are case sensitive.
	// Capitalizing the first character of object names to form correct URL.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(obj)
	if err != nil {
		return nil, err
	}

	// Adds the fields requirement parameter.
	fields := strings.Join(config.Fields.List(), ",")
	url.WithQueryParam("fields", fields)

	return url, nil
}
