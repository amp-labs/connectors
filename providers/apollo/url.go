package apollo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var restAPIPrefix string = "v1"

func (c *Connector) getURL(params common.ReadParams) (*urlbuilder.URL, error) {
	link, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	//If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		link.WithQueryParam("page", params.NextPage.String())
	}

	return link, nil
}
