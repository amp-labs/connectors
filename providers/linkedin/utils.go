package linkedin

import (
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	pageSize  = 100
	countSize = 100
)

var ObjectsWithSearchQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
	"dmpSegments",
	"adAnalytics",
)

var ObjectWithAccountId = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
)

// AccountIdInURLPathAndQueryParam defines the set of LinkedIn objects
// that require the `adAccountId` to be included both in the URL path
// (e.g., /rest/adAccounts/{adAccountId}/adCampaigns)
// and as a query parameter (dmpSegments?account={adAccountId}).
//
// These objects are ad-related and LinkedIn's API expects the account
// context in both the request path and query to properly scope results.
var accountIdInURLPathAndQueryParam = datautils.NewSet( //nolint:gochecknoglobals
	"adCampaignGroups",
	"adCampaigns",
	"dmpSegments",
	"adAnalytics",
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
			return handleCursorPagination(node)
		}

		if normalPaginationObject.Has(objName) {
			return handleNormalPagination(node)
		}

		return "", nil
	}
}

func handleCursorPagination(node *ajson.Node) (string, error) {
	pagination, err := jsonquery.New(node).ObjectOptional("metadata")
	if err != nil {
		return "", err
	}

	if pagination != nil {
		nextPage, err := jsonquery.New(pagination).StrWithDefault("nextPageToken", "")
		if err != nil {
			return "", err
		}

		if nextPage != "" {
			return nextPage, nil
		}
	}

	return "", nil
}

func handleNormalPagination(node *ajson.Node) (string, error) {
	paging, err := jsonquery.New(node).ObjectOptional("paging")
	if err != nil {
		return "", err
	}

	if paging != nil {
		nextPage, err := jsonquery.New(paging).IntegerWithDefault("count", 0)
		if err != nil {
			return "", err
		}

		if nextPage != 0 {
			start, err := jsonquery.New(paging).IntegerWithDefault("start", 0)
			if err != nil {
				return "", err
			}

			return strconv.Itoa(int(start) + int(nextPage)), nil
		}
	}

	return "", nil
}

//nolint:cyclop,funlen
func (c *Connector) buildReadURL(params common.ReadParams) (string, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return "", err
	}

	if ObjectsWithSearchQueryParam.Has(params.ObjectName) {
		switch params.ObjectName {
		case "dmpSegments":
			// nolint:lll
			// https://learn.microsoft.com/en-us/linkedin/marketing/matched-audiences/create-and-manage-segments?tabs=http#find-dmp-segments-by-account
			url.WithQueryParam("q", "account")

			url.WithQueryParam("start", "0")

			url.WithQueryParam("count", strconv.Itoa(countSize))

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

			url.WithQueryParam("pageSize", strconv.Itoa(pageSize))
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

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	// Returns an error if the object is ad-related and the adAccountId is missing.
	// It doesn't affect non-ads related object.
	if accountIdInURLPathAndQueryParam.Has(objName) && c.AdAccountId == "" {
		return nil, fmt.Errorf("missing adAccountId: this object (%s) requires an ad account ID", objName)
	}

	switch {
	case ObjectWithAccountId.Has(objName):
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", "adAccounts", c.AdAccountId, objName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", objName)
	}

	if err != nil {
		return nil, err
	}

	return url, nil
}
