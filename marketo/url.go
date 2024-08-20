package marketo

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var restAPIPrefix string = "rest" //nolint:gochecknoglobals

func (c *Connector) getURL(params common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(params.NextPage) > 0 {
		return constructURL(params.NextPage.String())
	}

	link, err := c.getApiURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// This affects  a very few number of objects.
	// Leads, Deleted Leads, Lead Changes,
	if !params.Since.IsZero() {
		time := params.Since.Format(time.RFC3339)
		fmtTime := fmt.Sprintf("%v", time)
		link.WithQueryParam("sinceDatetime", fmtTime)
	}

	return link, nil
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	return link, nil
}
