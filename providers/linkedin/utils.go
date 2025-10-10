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

var cursorPaginationObject = datautils.NewSet( //nolint:gochecknoglobals
	"adAccounts",
	"adCampaignGroups",
	"adCampaigns",
)

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
			start, err := jsonquery.New(paging).IntegerWithDefault("start	", 0)
			if err != nil {
				return "", err
			}

			return strconv.Itoa(int(start) + int(nextPage)), nil
		}
	}

	return "", nil
}

//nolint:cyclop
func (c *Connector) buildReadURL(params common.ReadParams) (string, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	switch {
	case ObjectWithAccountId.Has(params.ObjectName):
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", "adAccounts", c.AdAccountId, params.ObjectName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", params.ObjectName)
	}

	if err != nil {
		return "", err
	}

	if ObjectsWithSearchQueryParam.Has(params.ObjectName) {
		switch params.ObjectName {
		case "dmpSegments":
			url.WithQueryParam("q", "account")

			url.WithQueryParam("start", "0")

			url.WithQueryParam("count", strconv.Itoa(countSize))

			accountsValue := fmt.Sprintf("urn%%3Ali%%3AsponsoredAccount%%3A%s", c.AdAccountId) //nolint:perfsprint

			url.WithUnencodedQueryParam("account", accountsValue)
		case "adAnalytics":
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
