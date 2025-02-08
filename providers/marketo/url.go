package marketo

import (
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIPrefix = "rest" //nolint:gochecknoglobals

func (c *Connector) constructURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if err := constructURLQueries(url, params); err != nil {
		return nil, err
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	if !params.Since.IsZero() {
		switch c.Module.ID {
		case ModuleAssets:
			fmtTime := params.Since.Format(time.RFC3339)
			url.WithQueryParam("earliestUpdatedAt", fmtTime)
			url.WithQueryParam("latestUpdatedAt", time.Now().Format(time.RFC3339))

		default: // we currently don't support filtering in leads.
		}
	}

	return url, nil
}

func addFilteringIDQueries(urlbuilder *urlbuilder.URL, startIdx string) error {
	ids := make([]string, batchSize)

	idx, err := strconv.Atoi(startIdx)
	if err != nil {
		return err
	}

	for i := range ids {
		ids[i] = strconv.Itoa(idx + i)
	}

	queryIDs := strings.Join(ids, ",")
	urlbuilder.WithQueryParam("filterValues", queryIDs)
	urlbuilder.WithQueryParam("filterType", "id")

	return nil
}

func constructURLQueries(url *urlbuilder.URL, params common.ReadParams) error {
	if filtersByIDs(params.ObjectName) && len(params.NextPage) == 0 {
		if err := addFilteringIDQueries(url, "1"); err != nil {
			return err
		}
	}

	// If NextPage is set, then we're reading the next page of results.
	if len(params.NextPage) > 0 {
		if filtersByIDs(params.ObjectName) {
			if err := addFilteringIDQueries(url, params.NextPage.String()); err != nil {
				return err
			}
		} else {
			url.WithQueryParam("nextPageToken", params.NextPage.String())
		}
	}

	return nil
}
