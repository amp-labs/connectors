package salesloft

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.MakeRecordsFunc("data"),
		makeNextRecordsURL(url),
		common.MakeMarshaledDataFunc(flattenCustomEmbed),
		params.Fields,
	)
}

// objectsWithCursorPagination lists objects that support cursor-based pagination.
// These objects have an updated_at field AND the API sorts by updated_at when sort_direction is set.
// All other objects fall back to offset-based pagination (page=N).
//
//nolint:gochecknoglobals
var objectsWithCursorPagination = map[string]bool{
	"accounts":            true,
	"actions":             true,
	"activities/calls":    true,
	"activities/emails":   true,
	"cadence_memberships": true,
	"cadences":            true,
	"call_data_records":   true,
	"conversations":       true,
	"crm_activities":      true,
	"email_templates":     true,
	"notes":               true,
	"opportunities":       true,
	"opportunity_people":  true,
	"opportunity_stages":  true,
	"people":              true,
	"steps":               true,
	"successes":           true,
	"team_templates":      true,
}

func supportsCursorPagination(objectName string) bool {
	return objectsWithCursorPagination[objectName]
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getObjectURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	if supportsCursorPagination(config.ObjectName) {
		// Use cursor-based polling as recommended by Salesloft for efficient data retrieval.
		// Results are sorted by updated_at ascending so we can use the last record's timestamp
		// as the cursor for the next request, avoiding deep pagination (page 500+) which causes
		// rate limit cost escalation and server errors.
		// See: https://developers.salesloft.com/docs/platform/guides/building-an-efficient-cursor-poller/
		url.WithQueryParam("sort_by", "updated_at")
		url.WithQueryParam("sort_direction", "asc")

		if !config.Since.IsZero() {
			updatedSince := config.Since.Format(time.RFC3339Nano)
			url.WithQueryParam("updated_at[gte]", updatedSince)
		}
	} else if !config.Since.IsZero() {
		updatedSince := config.Since.Format(time.RFC3339Nano)
		url.WithQueryParam("updated_at[gte]", updatedSince)
	}

	return url, nil
}
