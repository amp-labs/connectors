package crm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	personsObjectName       = "persons"
	organizationsObjectName = "organizations"

	// Filter field names.
	searchFieldID    = "id"
	searchFieldEmail = "email"
	searchFieldPhone = "phone"

	searchPathSuffix = "search"

	// maxSearchLimit is the page size cap of a search.
	//
	// The search endpoint itself allows 500, but its matches are hydrated through the
	// `ids` query parameter of the list endpoint, which takes at most 100 identifiers.
	maxSearchLimit = 100
)

var (
	// ErrUnsupportedSearchField is returned when a filter names a field the object cannot be looked up by.
	ErrUnsupportedSearchField = errors.New("field is not supported in search filter")

	// ErrUnsupportedSearchOperator is returned for any operator other than equality.
	ErrUnsupportedSearchOperator = errors.New("unsupported search operator")

	// ErrUnsupportedSearchFilterCombination is returned when more than one field filter is given.
	ErrUnsupportedSearchFilterCombination = errors.New("unsupported combination of search filters")

	// ErrInvalidSearchValue is returned when a filter value is empty or of an unusable type.
	ErrInvalidSearchValue = errors.New("invalid search filter value")
)

// searchTermFields maps a filter field onto the value Pipedrive expects in the `fields`
// query parameter of the search endpoint, which narrows where the term is looked for.
var searchTermFields = datautils.Map[string, string]{ // nolint:gochecknoglobals
	searchFieldEmail: "email",
	searchFieldPhone: "phone",
}

// searchableFields lists the fields each object can be looked up by.
var searchableFields = datautils.Map[string, datautils.StringSet]{ // nolint:gochecknoglobals
	personsObjectName:       datautils.NewStringSet(searchFieldID, searchFieldEmail, searchFieldPhone),
	organizationsObjectName: datautils.NewStringSet(searchFieldID),
}

// Search looks up records by an exact value of one of their fields.
//
// Persons can be found by `id`, `email` and `phone`. Organizations by `id`.
// A search is always by a single field.
//
// Lookups by `id` are served by the list endpoint (`ids` query parameter).
// Lookups by `email` and `phone` go to the search endpoint with `exact_match` enabled. That endpoint
// returns trimmed down items, therefore every match is then fetched in full from the list endpoint,
// so that the records have the very same shape as the ones produced by Read.
//
// Phone matching is literal. A number stored as `77051838136` is found only by that exact string:
// `+77051838136`, `7705 183 8136` and a partial number all return nothing. The caller must send
// the phone exactly as it is stored in Pipedrive. (Emails, in contrast, match case-insensitively,
// but still by the whole value.)
//
// https://developers.pipedrive.com/docs/api/v1/Persons#searchPersons
// https://developers.pipedrive.com/docs/api/v1/Persons#getPersons
// https://developers.pipedrive.com/docs/api/v1/Organizations#getOrganizations
func (a *Adapter) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if params.RawFilter != nil {
		return nil, fmt.Errorf("%w: raw filters are not supported", common.ErrInvalidRawFilter)
	}

	url, viaSearchEndpoint, err := a.buildSearchURL(params)
	if err != nil {
		return nil, err
	}

	if viaSearchEndpoint {
		return a.searchAndHydrate(ctx, params, url)
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(data),
		nextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

// searchAndHydrate queries the search endpoint and then loads the full records of the matches.
// The next page token always refers to the search endpoint.
func (a *Adapter) searchAndHydrate(
	ctx context.Context, params *common.SearchParams, searchURL *urlbuilder.URL,
) (*common.SearchResult, error) {
	resp, err := a.Client.Get(ctx, searchURL.String())
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	identifiers, err := extractSearchItemIDs(body)
	if err != nil {
		return nil, err
	}

	nextPage, err := nextRecordsURL(searchURL)(body)
	if err != nil {
		return nil, err
	}

	if len(identifiers) == 0 {
		return &common.SearchResult{
			Rows: 0,
			Data: []common.ReadResultRow{},
			Done: true,
		}, nil
	}

	listURL, err := a.getAPIURL(readAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	listURL.WithQueryParam("ids", strings.Join(identifiers, ","))
	listURL.WithQueryParam("limit", strconv.Itoa(len(identifiers)))

	listResp, err := a.Client.Get(ctx, listURL.String())
	if err != nil {
		return nil, err
	}

	result, err := common.ParseResult(listResp,
		common.ExtractOptionalRecordsFromPath(data),
		func(*ajson.Node) (string, error) { return nextPage, nil },
		common.GetMarshaledData,
		params.Fields,
	)
	if err != nil {
		return nil, err
	}

	// A match may have been removed between the two calls, or be invisible to the list endpoint.
	// An emptied page must not stop the pagination of the search.
	if nextPage != "" {
		result.NextPage = common.NextPageToken(nextPage)
		result.Done = false
	}

	return result, nil
}

// buildSearchURL returns the URL to call and whether it targets the search endpoint.
func (a *Adapter) buildSearchURL(params *common.SearchParams) (*urlbuilder.URL, bool, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, false, err
		}

		return url, isSearchEndpoint(params.NextPage.String()), nil
	}

	field, value, err := resolveSearchFilter(params)
	if err != nil {
		return nil, false, err
	}

	url, err := a.getAPIURL(readAPIVersion, params.ObjectName)
	if err != nil {
		return nil, false, err
	}

	termField, viaSearchEndpoint := searchTermFields[field]
	if !viaSearchEndpoint {
		// Lookup by identifier is served by the list endpoint.
		url.WithQueryParam("ids", value)

		return url, false, nil
	}

	url.AddPath(searchPathSuffix)
	url.WithQueryParam("term", value)
	url.WithQueryParam("fields", termField)
	url.WithQueryParam("exact_match", "true")

	limit := int64(maxSearchLimit)
	if params.Limit > 0 && params.Limit < maxSearchLimit {
		limit = params.Limit
	}

	url.WithQueryParam("limit", strconv.FormatInt(limit, 10))

	return url, true, nil
}

// resolveSearchFilter validates the filter and returns the field to look up by and the value to look for.
// Every lookup is by a single field: neither endpoint can match on several of them at once.
func resolveSearchFilter(params *common.SearchParams) (field, value string, err error) {
	supported, ok := searchableFields[params.ObjectName]
	if !ok {
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

	value, err = searchValueToString(filter.Value)
	if err != nil {
		return "", "", fmt.Errorf("%w: field %q", err, filter.FieldName)
	}

	return filter.FieldName, value, nil
}

func searchValueToString(value any) (string, error) {
	var text string

	switch typed := value.(type) {
	case string:
		text = strings.TrimSpace(typed)
	case float64:
		text = strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		text = strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		text = fmt.Sprint(typed)
	case fmt.Stringer: // json.Number among others.
		text = typed.String()
	default:
		return "", fmt.Errorf("%w: unsupported type %T", ErrInvalidSearchValue, value)
	}

	if text == "" {
		return "", fmt.Errorf("%w: empty value", ErrInvalidSearchValue)
	}

	return text, nil
}

// isSearchEndpoint reports whether the URL (a next page token) points to the search endpoint.
func isSearchEndpoint(rawURL string) bool {
	path, _, _ := strings.Cut(rawURL, "?")

	return strings.HasSuffix(strings.TrimSuffix(path, "/"), "/"+searchPathSuffix)
}

// extractSearchItemIDs reads the identifiers of the matches from the search response.
//
//	{"data": {"items": [{"result_score": 0.5, "item": {"id": 1, "type": "person", ...}}]}}
func extractSearchItemIDs(body *ajson.Node) ([]string, error) {
	items, err := jsonquery.New(body, data).ArrayOptional("items")
	if err != nil {
		return nil, err
	}

	identifiers := make([]string, 0, len(items))

	for _, entry := range items {
		identifier, err := jsonquery.New(entry, "item").IntegerRequired(searchFieldID)
		if err != nil {
			return nil, err
		}

		identifiers = append(identifiers, strconv.FormatInt(identifier, 10))
	}

	return identifiers, nil
}
