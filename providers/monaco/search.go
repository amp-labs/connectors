package monaco

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/monaco/metadata"
)

// conditionEquals is what common.FilterOperatorEQ maps onto.
//
// Monaco also accepts `is`, and its own examples use both against what look
// like the same kinds of field -- sequences filters contact_id with `equals`
// while meetings filters account_id with `is`, both UUIDs. Which operators a
// given field actually accepts is reported per-field by
// GET /v1/schemas/{entity} under `allowed_operators`, and we have no API key to
// read it. `equals` is chosen because it is the operator the spec uses for the
// plainest equality example (contacts by email). If a field turns out to accept
// only `is`, searching on it will 400 and this mapping has to become per-field.
const conditionEquals = "equals"

// ErrUnsupportedSearchOperator is returned for any operator the framework may
// grow that Monaco has no equivalent for.
var ErrUnsupportedSearchOperator = errors.New("unsupported search operator")

// searchableObjects are the objects whose list request schema carries a
// `filters` property.
//
// This is the read set of POST-list objects minus audiences, whose
// AudienceListRequest exposes only page and page_size. The three GET-list
// objects (tags, users, sequenceTemplates) have no request body at all and so
// cannot be filtered either.
//
//nolint:gochecknoglobals
var searchableObjects = datautils.NewStringSet(
	objectAccounts,
	objectCampaigns,
	objectContacts,
	objectMeetings,
	objectOpportunities,
	objectSequences,
	objectTasks,
)

// Search runs a filtered list against the same POST /v1/<object>/list endpoint
// Read uses -- in Monaco the list endpoint *is* the search endpoint, and the
// only difference is whether `filters` is populated. Pagination, the response
// envelope and record extraction are therefore shared with Read verbatim.
//
// A flat filters array is implicitly AND-joined by Monaco, which matches the
// contract of common.SearchFilter.
func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !searchableObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	body, err := buildSearchRequestBody(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Post(ctx, endpointURL, body)
	if err != nil {
		return nil, err
	}

	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		common.MakeRecordsFunc(recordsKey),
		nextPageFromPagination,
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

func buildSearchRequestBody(params *common.SearchParams) (map[string]any, error) {
	page, err := resolvePage(params.NextPage)
	if err != nil {
		return nil, err
	}

	filters, err := buildSearchFilters(params.Filter)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		pageKey:     page,
		pageSizeKey: clampPageSize(int(params.Limit)),
		filtersKey:  filters,
	}, nil
}

func buildSearchFilters(filter common.SearchFilter) ([]map[string]any, error) {
	rules := make([]map[string]any, 0, len(filter.FieldFilters))

	for _, fieldFilter := range filter.FieldFilters {
		condition, err := searchCondition(fieldFilter.Operator)
		if err != nil {
			return nil, err
		}

		rules = append(rules, map[string]any{
			"field":     fieldFilter.FieldName,
			"condition": condition,
			"value":     fieldFilter.Value,
		})
	}

	return rules, nil
}

// searchCondition maps a framework operator onto a Monaco condition. The
// framework currently defines exactly one (FilterOperatorEQ), so Monaco's
// richer vocabulary -- contains, greater_than, less_than, is -- is not
// reachable through SearchParams today.
func searchCondition(operator common.FilterOperator) (string, error) {
	switch operator {
	case common.FilterOperatorEQ:
		return conditionEquals, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedSearchOperator, operator)
	}
}
