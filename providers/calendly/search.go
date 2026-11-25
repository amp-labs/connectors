package calendly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	scheduledEvents = "scheduled_events"

	// Filter field names. Each is named the way Calendly names it.
	searchFieldUUID         = "uuid"
	searchFieldInviteeEmail = "invitee_email"
	searchFieldStatus       = "status"
	searchFieldMinStartTime = "min_start_time"
	searchFieldMaxStartTime = "max_start_time"

	singleRecordKey = "resource"
	maxSearchLimit  = 100
)

var (
	// ErrUnsupportedSearchField is returned when a filter names a field the object cannot be looked up by.
	ErrUnsupportedSearchField = errors.New("field is not supported in search filter")

	// ErrUnsupportedSearchOperator is returned for any operator other than equality.
	ErrUnsupportedSearchOperator = errors.New("unsupported search operator")

	// ErrUnsupportedSearchFilterCombination is returned when the field filters cannot be used together.
	ErrUnsupportedSearchFilterCombination = errors.New("unsupported combination of search filters")

	// ErrInvalidSearchValue is returned when a filter value is empty or is of an unexpected type.
	ErrInvalidSearchValue = errors.New("invalid search filter value")
)

// searchableFields lists the fields each object can be looked up by.
var searchableFields = datautils.Map[string, datautils.StringSet]{ // nolint:gochecknoglobals
	scheduledEvents: datautils.NewStringSet(
		searchFieldUUID,
		searchFieldInviteeEmail,
		searchFieldStatus,
		searchFieldMinStartTime,
		searchFieldMaxStartTime,
	),
}

// Search looks up scheduled events by exact values of their fields.
//
// An event can be found by `uuid`, which is the last segment of the event `uri`. This filter must be
// used alone. It is served by the "get event" endpoint. An event that does not exist produces
// an empty result.
//
// Events can also be listed by any combination of the following filters, a field at most once:
//   - `invitee_email`: the whole address, matched case-insensitively. A partial address is rejected.
//   - `status`: `active` or `canceled`.
//   - `min_start_time`, `max_start_time`: the range the start time of an event falls into.
//     Both bounds are inclusive. Search knows the equality operator only, therefore the range is
//     expressed by these two fields, either or both. The value is a string in the RFC 3339 format
//     (`2026-12-01T06:00:00Z`) or time.Time.
//
// The list is limited to the events of the connected user, the same way Read is.
//
// https://developer.calendly.com/api-docs/e2f95ebd44914-get-event
// https://developer.calendly.com/api-docs/2d5ed9bbd2952-list-events
func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if params.RawFilter != nil {
		return nil, fmt.Errorf("%w: raw filters are not supported", common.ErrInvalidRawFilter)
	}

	if params.NextPage != "" {
		// Only the list of events is paginated. The next page URL carries the filters.
		return c.searchList(ctx, params, params.NextPage.String())
	}

	filters, err := resolveSearchFilters(params)
	if err != nil {
		return nil, err
	}

	if uuid, ok := filters[searchFieldUUID]; ok {
		return c.searchSingleRecord(ctx, params, uuid)
	}

	searchURL, err := c.buildSearchListURL(params, filters)
	if err != nil {
		return nil, err
	}

	return c.searchList(ctx, params, searchURL.String())
}

// searchList queries an endpoint which responds with a page of records.
func (c *Connector) searchList(
	ctx context.Context, params *common.SearchParams, rawURL string,
) (*common.SearchResult, error) {
	resp, err := c.JSONHTTPClient().Get(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractRecordsFromPath(dataKey),
		nextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

// searchSingleRecord fetches one record by its identifier.
// Missing record is not an error, the result is then empty.
func (c *Connector) searchSingleRecord(
	ctx context.Context, params *common.SearchParams, identifier string,
) (*common.SearchResult, error) {
	recordURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName, url.PathEscape(identifier))
	if err != nil {
		return nil, err
	}

	result := &common.SearchResult{
		Rows: 0,
		Data: []common.ReadResultRow{},
		Done: true,
	}

	resp, err := c.JSONHTTPClient().Get(ctx, recordURL.String())
	if err != nil {
		if httpErr, isHTTP := errors.AsType[*common.HTTPError](err); isHTTP && httpErr.Status == http.StatusNotFound {
			return result, nil
		}

		return nil, err
	}

	body, hasBody := resp.Body()
	if !hasBody {
		return result, nil
	}

	resource, err := jsonquery.New(body).ObjectRequired(singleRecordKey)
	if err != nil {
		return nil, err
	}

	record, err := jsonquery.Convertor.ObjectToMap(resource)
	if err != nil {
		return nil, err
	}

	result.Data, err = common.GetMarshaledData([]map[string]any{record}, params.Fields.List())
	if err != nil {
		return nil, err
	}

	result.Rows = int64(len(result.Data))

	return result, nil
}

func (c *Connector) buildSearchListURL(
	params *common.SearchParams, filters map[string]string,
) (*urlbuilder.URL, error) {
	searchURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, params.ObjectName)
	if err != nil {
		return nil, err
	}

	limit := int64(maxSearchLimit)
	if params.Limit > 0 && params.Limit < maxSearchLimit {
		limit = params.Limit
	}

	searchURL.WithQueryParam(countQuery, strconv.FormatInt(limit, 10))

	// The same scope Read applies.
	if requiresOrgURIQueryParam.Has(params.ObjectName) {
		searchURL.WithQueryParam(organization, c.orgURI)
	}

	if requiresUserURIQueryParam.Has(params.ObjectName) {
		searchURL.WithQueryParam(user, c.userURI)
	}

	// Filter fields carry the names of the query parameters.
	for field, value := range filters {
		searchURL.WithQueryParam(field, value)
	}

	return searchURL, nil
}

// resolveSearchFilters validates the filters and returns the values to look for keyed by the field name.
func resolveSearchFilters(params *common.SearchParams) (map[string]string, error) {
	supported, searchable := searchableFields[params.ObjectName]
	if !searchable {
		return nil, common.ErrOperationNotSupportedForObject
	}

	filters := make(map[string]string)

	for _, filter := range params.Filter.FieldFilters {
		if filter.Operator != common.FilterOperatorEQ {
			return nil, fmt.Errorf("%w: %q", ErrUnsupportedSearchOperator, filter.Operator)
		}

		if !supported.Has(filter.FieldName) {
			return nil, fmt.Errorf("%w: %q on %q", ErrUnsupportedSearchField, filter.FieldName, params.ObjectName)
		}

		if _, repeated := filters[filter.FieldName]; repeated {
			return nil, fmt.Errorf("%w: field %q is given more than once",
				ErrUnsupportedSearchFilterCombination, filter.FieldName)
		}

		value, err := searchValueToString(filter)
		if err != nil {
			return nil, err
		}

		filters[filter.FieldName] = value
	}

	if _, byUUID := filters[searchFieldUUID]; byUUID && len(filters) > 1 {
		return nil, fmt.Errorf("%w: field %q must be the only filter",
			ErrUnsupportedSearchFilterCombination, searchFieldUUID)
	}

	return filters, nil
}

func searchValueToString(filter common.FieldFilter) (string, error) {
	var text string

	switch value := filter.Value.(type) {
	case string:
		text = strings.TrimSpace(value)
	case time.Time:
		text = value.UTC().Format(time.RFC3339Nano)
	default:
		return "", fmt.Errorf("%w: field %q expects a string, got %T",
			ErrInvalidSearchValue, filter.FieldName, filter.Value)
	}

	if text == "" {
		return "", fmt.Errorf("%w: field %q is empty", ErrInvalidSearchValue, filter.FieldName)
	}

	return text, nil
}
