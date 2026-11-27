package front

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	conversations = "conversations"

	// Filter field names.
	searchFieldID     = "id"
	searchFieldEmail  = "email"
	searchFieldPhone  = "phone"
	searchFieldStatus = "status"

	maxSearchLimit = 100
)

var (
	// ErrUnsupportedSearchField is returned when a filter names a field the object cannot be looked up by.
	ErrUnsupportedSearchField = errors.New("field is not supported in search filter")

	// ErrUnsupportedSearchOperator is returned for any operator other than equality.
	ErrUnsupportedSearchOperator = errors.New("unsupported search operator")

	// ErrUnsupportedSearchFilterCombination is returned when more than one field filter is given.
	ErrUnsupportedSearchFilterCombination = errors.New("unsupported combination of search filters")

	// ErrInvalidSearchValue is returned when a filter value is empty or is not a string.
	ErrInvalidSearchValue = errors.New("invalid search filter value")
)

// searchableFields lists the fields each object can be looked up by.
var searchableFields = datautils.Map[string, datautils.StringSet]{ // nolint:gochecknoglobals
	contacts:      datautils.NewStringSet(searchFieldID, searchFieldEmail, searchFieldPhone),
	conversations: datautils.NewStringSet(searchFieldID, searchFieldStatus),
}

// contactAliasPrefixes are the resource aliases a contact can be fetched by instead of its identifier.
// https://dev.frontapp.com/docs/resource-aliases-1
var contactAliasPrefixes = datautils.Map[string, string]{ // nolint:gochecknoglobals
	searchFieldEmail: "alt:email:",
	searchFieldPhone: "alt:phone:",
}

// Search looks up records by an exact value of one of their fields. A search is always by a single field.
//
// Contacts can be found by `id`, `email` and `phone`. Every one of them is served by the
// "get contact" endpoint, the last two via a resource alias (`alt:email:...`, `alt:phone:...`).
// A contact that does not exist produces an empty result. Front is lenient about the phone format:
// a number stored as `+15555550100` is also found by `15555550100`, `5555550100` and `+1 555 555 0100`.
// Emails match case-insensitively.
//
// Conversations can be found by `id` and `status`. The former is served by the "get conversation"
// endpoint. The latter by the conversation search with the `is:<status>` query. The statuses known
// to the search are: open, archived, assigned, unassigned, unreplied, snoozed, trashed, waiting, resolved.
// The search endpoint has a stricter rate limit than the rest of the API.
//
// https://dev.frontapp.com/reference/get-contact-by-id
// https://dev.frontapp.com/reference/get-conversation-by-id
// https://dev.frontapp.com/reference/search-conversations
func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if params.RawFilter != nil {
		return nil, fmt.Errorf("%w: raw filters are not supported", common.ErrInvalidRawFilter)
	}

	if params.NextPage != "" {
		// Only the conversation search is paginated.
		return c.searchList(ctx, params, params.NextPage.String())
	}

	field, value, err := resolveSearchFilter(params)
	if err != nil {
		return nil, err
	}

	if field == searchFieldStatus {
		searchURL, err := c.buildConversationSearchURL(params, value)
		if err != nil {
			return nil, err
		}

		return c.searchList(ctx, params, searchURL.String())
	}

	return c.searchSingleRecord(ctx, params, contactAliasPrefixes[field]+value)
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
		common.ExtractRecordsFromPath(dataResponseKey),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

// searchSingleRecord fetches one record by its identifier or resource alias.
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
		if httpErr, ok := errors.AsType[*common.HTTPError](err); ok && httpErr.Status == http.StatusNotFound {
			return result, nil
		}

		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return result, nil
	}

	record, err := jsonquery.Convertor.ObjectToMap(body)
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

func (c *Connector) buildConversationSearchURL(
	params *common.SearchParams, status string,
) (*urlbuilder.URL, error) {
	searchURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, conversations, "search", url.PathEscape("is:"+status))
	if err != nil {
		return nil, err
	}

	limit := int64(maxSearchLimit)
	if params.Limit > 0 && params.Limit < maxSearchLimit {
		limit = params.Limit
	}

	searchURL.WithQueryParam(pageSizeKey, strconv.FormatInt(limit, 10))

	return searchURL, nil
}

// resolveSearchFilter validates the filter and returns the field to look up by and the value to look for.
func resolveSearchFilter(params *common.SearchParams) (field, value string, err error) {
	supported, searchable := searchableFields[params.ObjectName]
	if !searchable {
		return "", "", common.ErrOperationNotSupportedForObject
	}

	if len(params.Filter.FieldFilters) != 1 {
		return "", "", fmt.Errorf("%w: exactly one field filter is expected, got %d",
			ErrUnsupportedSearchFilterCombination, len(params.Filter.FieldFilters))
	}

	filter := params.Filter.FieldFilters[0]

	if filter.Operator != common.FilterOperatorEQ {
		return "", "", fmt.Errorf("%w: %q", ErrUnsupportedSearchOperator, filter.Operator)
	}

	if !supported.Has(filter.FieldName) {
		return "", "", fmt.Errorf("%w: %q on %q", ErrUnsupportedSearchField, filter.FieldName, params.ObjectName)
	}

	text, isString := filter.Value.(string)
	if !isString {
		return "", "", fmt.Errorf("%w: field %q expects a string, got %T",
			ErrInvalidSearchValue, filter.FieldName, filter.Value)
	}

	text = strings.TrimSpace(text)

	if text == "" {
		return "", "", fmt.Errorf("%w: field %q is empty", ErrInvalidSearchValue, filter.FieldName)
	}

	// Whitespace would turn a status into a search query made of several terms.
	if filter.FieldName == searchFieldStatus && strings.ContainsAny(text, " \t\n") {
		return "", "", fmt.Errorf("%w: field %q must be a single word", ErrInvalidSearchValue, filter.FieldName)
	}

	return filter.FieldName, text, nil
}
