package marketo

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var restAPIPrefix = "rest" //nolint:gochecknoglobals

func (c *Connector) getURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		url.WithQueryParam("nextPageToken", params.NextPage.String())
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() {
		switch c.Module {
		case ModuleAssets.String():
			t := params.Since.Format(time.RFC3339)
			fmtTime := fmt.Sprintf("%v", t)
			url.WithQueryParam("earliestUpdatedAt", fmtTime)
			url.WithQueryParam("latestUpdatedAt", time.Now().Format(time.RFC3339))

		default: // we currently don't support filtering in leads.
		}
	}

	return url, nil
}
