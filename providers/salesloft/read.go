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

// objectsWithoutUpdatedAt lists objects that don't have an updated_at field
// and therefore cannot use cursor-based pagination. These fall back to offset-based pagination.
//
//nolint:gochecknoglobals
var objectsWithoutUpdatedAt = map[string]bool{
	"account_team_member_roles":              true,
	"account_types":                          true,
	"audit_reports":                          true,
	"calendar_availabilities":                true,
	"crm_account_team_members":               true,
	"crm_team_members_with_roles":            true,
	"custom_roles":                           true,
	"data_control/requests":                  true,
	"email_template_attachments":             true,
	"external/configurations":                true,
	"external/mappings":                      true,
	"groups":                                 true,
	"integrations/signals/registrations":       true,
	"integrations/signals/registrations/plays": true,
	"pending_emails":                         true,
	"phone_number_assignments":               true,
	"phone_numbers/caller_ids":               true,
	"saved_list_views":                       true,
	"tags":                                   true,
	"team_template_attachments":              true,
	"users":                                  true,
	"webhook_subscriptions":                  true,
}

func supportsCursorPagination(objectName string) bool {
	return !objectsWithoutUpdatedAt[objectName]
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
		url.WithQueryParam("sort_direction", "ASC")

		if !config.Since.IsZero() {
			updatedSince := config.Since.Format(time.RFC3339Nano)
			url.WithQueryParam("updated_at[gte]", updatedSince)
		}
	}

	return url, nil
}
