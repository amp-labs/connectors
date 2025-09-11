package calendly

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

/*
Objects supporting since query params:
- activity_log_entries - min_occurred_at, max_occurred_at - not necessary.
- user_busy_times - start_time, end_time - both are required
- outgoing_communications - min_created_at,max_created_at - not required
- scheduled_events - min_start_time, max_start_time - not required
*/

const (
	organization = "organization"
	user         = "user"

	activityLogEntries     = "activity_log_entries"
	userBusyTimes          = "user_busy_times"
	outgoingCommunications = "outgoing_communications"
	scheduledEvents        = "scheduled_events"
)

// when reading the objects "scheduled_events" and "user_busy_times" with since and until params provided
// This connector uses event start timestamps rather than create/update timestamps because they are calendar items,
// unlike other objects.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (string, error) { // nolint: cyclop
	var (
		url string
		err error
	)

	urlBuilder, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return "", err
	}

	if requiresOrgURIQueryParam.Has(params.ObjectName) {
		urlBuilder.WithQueryParam(organization, c.orgURI)
	}

	if requiresUserURIQueryParam.Has(params.ObjectName) {
		urlBuilder.WithQueryParam(user, c.userURI)
	}

	if !params.Since.IsZero() {
		switch params.ObjectName {
		case activityLogEntries:
			urlBuilder.WithQueryParam("min_occurred_at", params.Since.Format(time.RFC3339))
		case userBusyTimes:
			urlBuilder.WithQueryParam("start_time", params.Since.Format(time.RFC3339))
		case outgoingCommunications:
			urlBuilder.WithQueryParam("min_created_at", params.Since.Format(time.RFC3339))
		case scheduledEvents:
			urlBuilder.WithQueryParam("min_start_time", params.Since.Format(time.RFC3339))
		}
	}

	if !params.Until.IsZero() {
		switch params.ObjectName {
		case activityLogEntries:
			urlBuilder.WithQueryParam("max_occurred_at", params.Until.Format(time.RFC3339))
		case userBusyTimes:
			urlBuilder.WithQueryParam("end_time", params.Until.Format(time.RFC3339))
		case outgoingCommunications:
			urlBuilder.WithQueryParam("max_created_at", params.Until.Format(time.RFC3339))
		case scheduledEvents:
			urlBuilder.WithQueryParam("max_start_time", params.Until.Format(time.RFC3339))
		}
	}

	url = urlBuilder.String()

	if params.NextPage != "" {
		url = params.NextPage.String()
	}

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(dataKey),
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}
