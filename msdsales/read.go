package msdsales

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var annotationsHeader = common.Header{
	Key:   "Prefer",
	Value: `odata.include-annotations="*"`,
}

// Microsoft API supports other capabilities like filtering, grouping, and sorting which we can potentially tap into later.
// See https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/query-data-web-api#odata-query-options
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var fullURL string

	if len(config.NextPage) == 0 {
		// First page
		relativeURL := config.ObjectName + makeQueryValues(config)
		fullURL = c.getURL(relativeURL)
	} else {
		// Next page
		fullURL = config.NextPage.String()
	}
	rsp, err := c.get(ctx, fullURL, newPaginationHeader(DefaultPageSize), annotationsHeader)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		config.Fields,
	)
}

func newPaginationHeader(pageSize int) common.Header {
	return common.Header{
		Key:   "Prefer",
		Value: fmt.Sprintf("odata.maxpagesize=%v", pageSize),
	}
}

func makeQueryValues(config common.ReadParams) string {
	queryValues := url.Values{}

	if len(config.Fields) != 0 {
		queryValues.Add("$select", strings.Join(config.Fields, ","))
	}

	result := queryValues.Encode()
	if len(result) != 0 {
		// TODO this is a hack. net/url encodes $. But we rely heavily on it
		// same problem with "@" ex: @Microsoft.Dynamics.CRM.totalrecordcountlimitexceeded
		// @ symbol is removed
		for before, after := range map[string]string{
			"%24select": "$select",
		} {
			result = strings.ReplaceAll(result, before, after)
		}

		result = strings.ReplaceAll(result, "%40", "@")
		result = strings.ReplaceAll(result, "%2C", ",")

		return "?" + result
	}

	return result
}
