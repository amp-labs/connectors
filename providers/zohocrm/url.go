package zohocrm

import (
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// Check if we're reading the next-page.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	// Just incase someone sends leads, Instead of Leads
	// All Objects are capitalized in their API names.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(obj)
	if err != nil {
		return nil, err
	}

	// Adds the fields requirement parameter.
	fields := strings.Join(config.Fields.List(), ",")
	url.WithQueryParam("fields", fields)

	// This will be added during the first call.
	// This is a custom parameter, we add this in the request
	// So as we do not loose the state of the since parameter value.
	// `since` is just a preferred name parameter.
	if !config.Since.IsZero() {
		url.WithQueryParam("since", config.Since.Format(time.RFC3339))
	}

	return url, nil
}
