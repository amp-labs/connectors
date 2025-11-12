package ads

import (
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/linkedin/internal/shared"
	"github.com/spyzhov/ajson"
)

var objectsWithSearchQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
	"dmpSegments",
	"adAnalytics",
)

var objectWithAccountId = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
)

// cursorPaginationObject holds the list of objects that use cursor-based pagination.
// These endpoints require passing a cursor (like "nextPageToken") to fetch paginated results.
var cursorPaginationObject = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
)

// normalPaginationObject holds the list of objects that use offset-based (normal) pagination.
// These endpoints use standard pagination parameters like "count" and "start".
var normalPaginationObject = datautils.NewSet( //nolint:gochecknoglobals
	"dmpSegments",
)

// For dmpSegment follow offset like pagination remaining object follows cursor pagination.
func makeNextRecord(objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if cursorPaginationObject.Has(objName) {
			return shared.HandleCursorPagination(node)
		}

		if normalPaginationObject.Has(objName) {
			return shared.HandleOffsetPagination(node)
		}

		return "", nil
	}
}

//nolint:cyclop,funlen
func (c *Adapter) buildReadURL(params common.ReadParams) (string, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return "", err
	}

	if objectsWithSearchQueryParam.Has(params.ObjectName) {
		switch params.ObjectName {
		case "dmpSegments":
			// nolint:lll
			// https://learn.microsoft.com/en-us/linkedin/marketing/matched-audiences/create-and-manage-segments?tabs=http#find-dmp-segments-by-account
			url.WithQueryParam("q", "account")

			url.WithQueryParam("start", "0")

			url.WithQueryParam("count", strconv.Itoa(shared.CountSize))

			accountsValue := fmt.Sprintf("urn%%3Ali%%3AsponsoredAccount%%3A%s", c.AdAccountId) //nolint:perfsprint

			url.WithUnencodedQueryParam("account", accountsValue)
		case "adAnalytics":
			//nolint:lll
			//https://learn.microsoft.com/en-us/linkedin/marketing/integrations/ads-reporting/ads-reporting?tabs=http#analytics-finder
			url.WithQueryParam("q", "analytics")

			url.WithQueryParam("timeGranularity", "DAILY")

			// Encode only the "urn" inside List(), leave parentheses literal
			accountsValue := fmt.Sprintf("List(urn%%3Ali%%3AsponsoredAccount%%3A%s)", c.AdAccountId) //nolint:perfsprint

			url.WithUnencodedQueryParam("accounts", accountsValue)

			// Handle dateRange manually
			var dateRange string

			if !params.Since.IsZero() {
				startDate := fmt.Sprintf("start:(year:%d,month:%d,day:%d)",
					params.Since.Year(), int(params.Since.Month()), params.Since.Day())

				endDate := ""

				if !params.Until.IsZero() {
					endDate = fmt.Sprintf(",end:(year:%d,month:%d,day:%d)",
						params.Until.Year(), int(params.Until.Month()), params.Until.Day())
				}

				dateRange = fmt.Sprintf("(%s%s)", startDate, endDate)
			}

			if dateRange != "" {
				url.WithUnencodedQueryParam("dateRange", dateRange)
			}
		default:
			url.WithQueryParam("q", "search")

			url.WithQueryParam("pageSize", strconv.Itoa(shared.PageSize))
		}
	}

	if len(params.NextPage) != 0 {
		if cursorPaginationObject.Has(params.ObjectName) {
			url.WithQueryParam("pageToken", params.NextPage.String())
		} else {
			url.WithQueryParam("start", params.NextPage.String())
		}
	}

	return url.String(), nil
}

func (c *Adapter) constructURL(objName string) (*urlbuilder.URL, error) {
	switch {
	case objectWithAccountId.Has(objName):
		return urlbuilder.New(c.ModuleInfo().BaseURL, "rest", "adAccounts", c.AdAccountId, objName)
	default:
		return urlbuilder.New(c.ModuleInfo().BaseURL, "rest", objName)
	}
}
