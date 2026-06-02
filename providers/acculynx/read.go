package acculynx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
	"github.com/spyzhov/ajson"
)

var errChildPagesExceeded = errors.New("acculynx: nested fetch exceeded page cap")

const (
	// AccuLynx OpenAPI does not document maximum pageSize, but its API enforces
	// server-side per-object caps: /jobs and /supplements reject pageSize > 25,
	// other list endpoints reject > 50. 25 is the strictest cap and is safe
	// across every object. Well within AccuLynx's 10 req/sec per-key limit.
	defaultPageSize = "25"
	maxPageSize     = 25

	pageSizeParam    = "pageSize"
	recordStartParam = "recordStartIndex"
	pageStartParam   = "pageStartIndex"
	pageNumberParam  = "pageNumber"

	// AccuLynx 10 req/sec per API key gives plenty of headroom; 4 is a
	// conservative cap for the per-parent fan-out.
	maxConcurrentChildFetch = 4

	// Safety cap on per-parent pagination — bounds the worst-case latency of
	// one Read call. 200 pages × 25 per page = 5000 records per parent,
	// generous for any sane window. For unbounded reads on outlier datasets
	// (e.g. jobs with hundreds of thousands of history entries), the cap fires
	// with a clear error pointing the caller at Since/Until — see fetchChildPages.
	maxChildPagesPerParent = 200

	// /calendars/{calendarId}/appointments requires startDate/endDate. When the
	// caller supplies neither, default to a 30-day window ending now.
	defaultAppointmentsWindow = 30 * 24 * time.Hour
)

const (
	objectJobs        = "jobs"
	objectContacts    = "contacts"
	objectEstimates   = "estimates"
	objectSupplements = "supplements"
	objectCalendars   = "calendars"
)

type nestedSpec struct {
	parentObject string
	leafSuffix   string
}

//nolint:gochecknoglobals
var nestedObjects = datautils.Map[string, nestedSpec]{
	"jobs/contacts":          {parentObject: objectJobs, leafSuffix: "contacts"},
	"jobs/custom-fields":     {parentObject: objectJobs, leafSuffix: "custom-fields"},
	"jobs/estimates":         {parentObject: objectJobs, leafSuffix: "estimates"},
	"jobs/history":           {parentObject: objectJobs, leafSuffix: "history"},
	"jobs/invoices":          {parentObject: objectJobs, leafSuffix: "invoices"},
	"jobs/milestone-history": {parentObject: objectJobs, leafSuffix: "milestone-history"},
	"jobs/representatives":   {parentObject: objectJobs, leafSuffix: "representatives"},

	"contacts/custom-fields":   {parentObject: objectContacts, leafSuffix: "custom-fields"},
	"contacts/email-addresses": {parentObject: objectContacts, leafSuffix: "email-addresses"},
	"contacts/phone-numbers":   {parentObject: objectContacts, leafSuffix: "phone-numbers"},

	"estimates/sections": {parentObject: objectEstimates, leafSuffix: "sections"},

	"supplements/items":     {parentObject: objectSupplements, leafSuffix: "items"},
	"supplements/notations": {parentObject: objectSupplements, leafSuffix: "notations"},

	"calendars/appointments": {parentObject: objectCalendars, leafSuffix: "appointments"},
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if _, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName); err != nil {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := c.buildInitialURL(params)
	if err != nil {
		return nil, err
	}

	applyJobsIncrementalFilter(url, params)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// buildInitialURL returns the first-page URL. For nested objects it returns
// the parent-list URL — the fan-out happens in parseReadResponse.
func (c *Connector) buildInitialURL(params common.ReadParams) (*urlbuilder.URL, error) {
	objectName := params.ObjectName
	if nested, isNested := nestedObjects[objectName]; isNested {
		objectName = nested.parentObject
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), objectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	applyPagination(url, objectName, params)

	return url, nil
}

func applyPagination(url *urlbuilder.URL, objectName string, params common.ReadParams) {
	spec := objectReadSpecs.Get(objectName)
	if spec.pagination == paginationNone {
		return
	}

	url.WithQueryParam(pageSizeParam, pageSizeWithCap(params))

	switch spec.pagination {
	case paginationOffsetRecord:
		url.WithQueryParam(recordStartParam, "0")
	case paginationOffsetPage:
		url.WithQueryParam(pageStartParam, "0")
	case paginationPageNumber:
		url.WithQueryParam(pageNumberParam, "1")
	case paginationNone:
		// No-op — handled above.
	}
}

// pageSizeWithCap returns params.PageSize when within bounds, otherwise
// defaultPageSize. AccuLynx's OpenAPI documents no hard maximum; maxPageSize
// matches the established convention across the repo.
func pageSizeWithCap(params common.ReadParams) string {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return defaultPageSize
	}

	return strconv.Itoa(params.PageSize)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if _, isNested := nestedObjects[params.ObjectName]; isNested {
		return c.parseNestedResponse(ctx, params, request, resp)
	}

	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	var transformer common.RecordTransformer

	if usesCustomFields(params.ObjectName) {
		transformer, err = c.buildCustomFieldsTransformer(ctx, params, resp)
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResultFiltered(
		params,
		resp,
		c.recordsFunc(params.ObjectName),
		c.makeFilterFunc(params, reqURL),
		common.MakeMarshaledDataFunc(transformer),
		params.Fields,
	)
}

// buildCustomFieldsTransformer fetches custom-field definitions and per-record
// values, returning a transformer that flattens those values onto the
// marshalled record. Caller must guarantee usesCustomFields(params.ObjectName)
// is true. When there is nothing to attach (empty body or no definitions for
// this entity), the returned transformer carries an empty map and short-
// circuits per record without touching the value-fan-out path.
func (c *Connector) buildCustomFieldsTransformer(
	ctx context.Context,
	params common.ReadParams,
	resp *common.JSONHTTPResponse,
) (common.RecordTransformer, error) {
	body, hasBody := resp.Body()
	if !hasBody {
		return attachReadCustomFields(nil), nil
	}

	defs, err := c.fetchCustomFieldDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	entity, hasEntity := customFieldEntityByObject[params.ObjectName]
	if !hasEntity || len(defs[entity]) == 0 {
		return attachReadCustomFields(nil), nil
	}

	parentIDs, err := c.extractParentIDsFromBody(params.ObjectName, body)
	if err != nil {
		return nil, err
	}

	values, err := c.fetchCustomFieldValuesForRecords(ctx, params.ObjectName, parentIDs)
	if err != nil {
		return nil, err
	}

	return attachReadCustomFields(values), nil
}

// recordsFunc resolves the records-array key from the schema's responseKey.
// Most AccuLynx list responses wrap as {..., items: [...]}; the exception is
// /acculynx/units-of-measure which uses "unitsOfMeasure".
func (c *Connector) recordsFunc(objectName string) common.NodeRecordsFunc {
	return common.MakeRecordsFunc(c.arrayFieldName(objectName))
}

func (c *Connector) arrayFieldName(objectName string) string {
	return metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)
}

// parseNestedResponse extracts parent IDs from the parent-list response, fans
// out one request per parent to the leaf endpoint, and returns the flattened
// records. Pagination advances through the parent list.
func (c *Connector) parseNestedResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.ReadResult{Done: true}, nil
	}

	nested := nestedObjects[params.ObjectName]

	parentNodes, err := c.recordsFunc(nested.parentObject)(body)
	if err != nil {
		return nil, err
	}

	rows, err := c.fetchChildrenForParents(ctx, parentNodes, params, nested)
	if err != nil {
		return nil, err
	}

	reqURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	nextPage, err := c.makeNextPage(nested.parentObject, reqURL)(body)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(rows)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     nextPage == "",
	}, nil
}

func extractIDs(nodes []*ajson.Node) []string {
	ids := make([]string, 0, len(nodes))

	for _, node := range nodes {
		id, err := jsonquery.New(node).StringRequired("id")
		if err != nil {
			continue
		}

		ids = append(ids, id)
	}

	return ids
}

// fetchChildrenForParents fans out one request per parent ID concurrently and
// preserves the parent order in the flattened result.
func (c *Connector) fetchChildrenForParents(
	ctx context.Context,
	parents []*ajson.Node,
	params common.ReadParams,
	nested nestedSpec,
) ([]common.ReadResultRow, error) {
	ids := extractIDs(parents)
	if len(ids) == 0 {
		return nil, nil
	}

	results := make([][]common.ReadResultRow, len(ids))
	jobs := make([]simultaneously.Job, len(ids))

	for i, parentID := range ids {
		idx, id := i, parentID

		jobs[idx] = func(ctx context.Context) error {
			rows, fetchErr := c.fetchChildPages(ctx, id, params, nested)
			if fetchErr != nil {
				return fmt.Errorf("fetching %s for %s %s: %w",
					nested.leafSuffix, nested.parentObject, id, fetchErr)
			}

			results[idx] = rows

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
		return nil, err
	}

	var data []common.ReadResultRow
	for _, rows := range results {
		data = append(data, rows...)
	}

	return data, nil
}

// fetchChildPages walks every page of one parent's nested collection.
// AccuLynx server-side paginates these endpoints (jobs/invoices,
// jobs/estimates, jobs/history, etc. — verified in the OpenAPI spec), so
// the loop follows result.NextPage until the records-array length signals the
// last page. Children configured as paginationNone exit after one iteration
// because makeNextPage returns "" immediately.
func (c *Connector) fetchChildPages(
	ctx context.Context,
	parentID string,
	params common.ReadParams,
	nested nestedSpec,
) ([]common.ReadResultRow, error) {
	parentPath, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), nested.parentObject)
	if err != nil {
		return nil, err
	}

	parentPath = strings.TrimSuffix(parentPath, "/")

	reqURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, parentPath, parentID, nested.leafSuffix)
	if err != nil {
		return nil, err
	}

	applyPagination(reqURL, params.ObjectName, params)
	applyHistoryDateWindow(reqURL, params)

	if params.ObjectName == "calendars/appointments" {
		applyAppointmentsDateWindow(reqURL, params)
	}

	var allRows []common.ReadResultRow

	for range maxChildPagesPerParent {
		resp, err := c.JSONHTTPClient().Get(ctx, reqURL.String())
		if err != nil {
			return nil, err
		}

		result, err := common.ParseResultFiltered(
			params,
			resp,
			c.recordsFunc(params.ObjectName),
			c.makeFilterFunc(params, reqURL),
			common.MakeMarshaledDataFunc(nil),
			params.Fields,
		)
		if err != nil {
			return nil, err
		}

		allRows = append(allRows, result.Data...)

		if result.NextPage == "" {
			return allRows, nil
		}

		reqURL, err = parseNextPageURL(result.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf(
		"%w: %s/%s/%s after %d pages — for jobs/history pass Since/Until to filter server-side",
		errChildPagesExceeded, nested.parentObject, parentID, nested.leafSuffix, maxChildPagesPerParent)
}

func parseNextPageURL(s string) (*urlbuilder.URL, error) {
	parsed, err := neturl.Parse(s)
	if err != nil {
		return nil, err
	}

	return urlbuilder.FromRawURL(parsed)
}

// applyAppointmentsDateWindow ensures /calendars/{id}/appointments always
// receives startDate + endDate (the OpenAPI marks both as required). Defaults
// to a 30-day window ending now when the caller supplies neither bound.
func applyAppointmentsDateWindow(url *urlbuilder.URL, params common.ReadParams) {
	until := params.Until
	if until.IsZero() {
		until = time.Now().UTC()
	}

	since := params.Since
	if since.IsZero() {
		since = until.Add(-defaultAppointmentsWindow)
	}

	url.WithQueryParam("startDate", since.Format(time.DateOnly))
	url.WithQueryParam("endDate", until.Format(time.DateOnly))
}

// makeNextPage returns a NextPageFunc tailored to the object's paginationStyle.
// Pagination terminates when the records array is shorter than pageSize (the
// universal "partial page = last page" signal used elsewhere in the repo).
func (c *Connector) makeNextPage(objectName string, reqURL *urlbuilder.URL) common.NextPageFunc {
	spec := objectReadSpecs.Get(objectName)
	recordsKey := c.arrayFieldName(objectName)

	return func(root *ajson.Node) (string, error) {
		if reqURL == nil || root == nil || spec.pagination == paginationNone {
			return "", nil
		}

		records, err := arrayLength(root, recordsKey)
		if err != nil {
			return "", err
		}

		per := queryParamIntOrDefault(reqURL, pageSizeParam, maxPageSize)
		if records < per {
			return "", nil
		}

		next, err := cloneURL(reqURL)
		if err != nil {
			return "", err
		}

		advancePagination(next, spec.pagination, records)

		return next.String(), nil
	}
}

func advancePagination(u *urlbuilder.URL, style paginationStyle, records int) {
	switch style {
	case paginationOffsetRecord:
		advanceOffset(u, recordStartParam, records)
	case paginationOffsetPage:
		advanceOffset(u, pageStartParam, records)
	case paginationPageNumber:
		advancePageNumber(u)
	case paginationNone:
		// No-op.
	}
}

func arrayLength(root *ajson.Node, recordsKey string) (int, error) {
	nodes, err := common.MakeRecordsFunc(recordsKey)(root)
	if err != nil {
		return 0, err
	}

	return len(nodes), nil
}

func advanceOffset(u *urlbuilder.URL, paramName string, increment int) {
	current, _ := u.GetFirstQueryParam(paramName)
	currentInt, _ := strconv.Atoi(current)
	u.WithQueryParam(paramName, strconv.Itoa(currentInt+increment))
}

func advancePageNumber(u *urlbuilder.URL) {
	current, _ := u.GetFirstQueryParam(pageNumberParam)

	currentInt, err := strconv.Atoi(current)
	if err != nil || currentInt < 1 {
		currentInt = 1
	}

	u.WithQueryParam(pageNumberParam, strconv.Itoa(currentInt+1))
}

func queryParamIntOrDefault(u *urlbuilder.URL, key string, defaultValue int) int {
	raw, _ := u.GetFirstQueryParam(key)

	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultValue
	}

	return v
}

func cloneURL(u *urlbuilder.URL) (*urlbuilder.URL, error) {
	return urlbuilder.New(u.String())
}
