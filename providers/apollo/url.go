package apollo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	restAPIPrefix string = "v1"  //nolint:gochecknoglobals
	pageSize      string = "100" //nolint:gochecknoglobals
)
var pageQuery string = "page" //nolint:gochecknoglobals

func (c *Connector) getURL(params common.ReadParams) (*urlbuilder.URL, error) {
	link, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		link.WithQueryParam(pageQuery, params.NextPage.String())
	}

	return link, nil
}
