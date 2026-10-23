package sybill

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/sybill/internal/metadata"
)

const (
	// Every list endpoint accepts a limit between 1 and 50.
	// https://api.sybill.ai/docs/pagination.html#controlling-page-size
	maxPageSize = 50
	// A conversation read that returns transcripts or summaries makes one detail request per record,
	// and every request counts toward 60 per minute per key, so its pages are kept small.
	// https://api.sybill.ai/docs/rate-limiting.html
	detailPageSize = 10

	// Conversations filter on meeting start time only; there is no updated-time filter.
	// https://api.sybill.ai/docs/endpoints.html#conversations
	startedAfterParam  = "startedAfter"
	startedBeforeParam = "startedBefore"

	// conversationLookback widens Since so a meeting is read again for some hours after it starts.
	// Transcript, summary and recordings are filled in asynchronously after the meeting ends, and
	// there is no status or updated-time field to ask for them, so a meeting first read while it is
	// in progress or still processing comes back with them later. The cost is that the same
	// conversation is delivered on every scheduled read within the window.
	// https://api.sybill.ai/docs/introduction.html#data-visibility-and-processing
	conversationLookback = 6 * time.Hour
)

// buildReadRequest lists records with cursor pagination. Filters and limit are sent on every page,
// as the cursor does not carry them. Accounts and deals have no updated-time filter (only created
// and last-activity dates, which don't change when a record is edited), so they are read in full.
// https://api.sybill.ai/docs/pagination.html#combining-with-filters
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	defaultPageSize := maxPageSize
	if needsConversationDetails(params) {
		defaultPageSize = detailPageSize
	}

	readURL.WithQueryParam("limit",
		strconv.Itoa(min(readhelper.PageSizeWithDefault(params, defaultPageSize), maxPageSize)))

	if params.ObjectName == objectNameConversations {
		if !params.Since.IsZero() {
			readURL.WithQueryParam(startedAfterParam,
				datautils.Time.FormatRFC3339inUTC(params.Since.Add(-conversationLookback)))
		}

		// startedBefore is exclusive (a live read leaves out a meeting starting exactly at it),
		// so one second is added to keep Until inclusive at the second precision of startTime.
		if !params.Until.IsZero() {
			readURL.WithQueryParam(startedBeforeParam,
				datautils.Time.FormatRFC3339inUTC(params.Until.Truncate(time.Second).Add(time.Second)))
		}
	}

	if params.NextPage != "" {
		readURL.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	idField := readhelper.NewIdField(idFields[params.ObjectName])
	marshal := readhelper.MakeMarshaledDataFuncWithId(nil, idField)

	if needsConversationDetails(params) {
		details, err := c.fetchConversationDetails(ctx, response)
		if err != nil {
			return nil, err
		}

		marshal = marshalConversationDetails(details)
	}

	return common.ParseResult(
		response,
		common.MakeRecordsFunc(params.ObjectName),
		nextPage,
		marshal,
		params.Fields,
	)
}
