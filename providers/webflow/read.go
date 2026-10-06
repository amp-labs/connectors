package webflow

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/webflow/metadata"
	"github.com/spyzhov/ajson"
)

// errMissingSiteID is returned when a site-scoped object is read without a
// site id. RequireMetadata makes this unreachable through NewConnector; it
// guards direct construction.
var errMissingSiteID = errors.New("webflow: a siteId is required to read this object")

const (
	// Webflow caps limit at 100 (spec: "max limit: 100"); larger values are
	// accepted but silently reduced, so cap here to keep pagination honest.
	defaultPageSize = 100
	maxPageSize     = 100

	limitParam     = "limit"
	offsetParam    = "offset"
	sortByParam    = "sortBy"
	sortOrderParam = "sortOrder"
	sortDescending = "desc"

	sitePlaceholder = "{site_id}"
	paginationKey   = "pagination"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	spec, ok := readSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Subsequent pages are fetched from the exact URL handed back as NextPage.
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := c.buildFirstPageURL(params, spec)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildFirstPageURL(params common.ReadParams, spec readSpec) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	path, err = c.substituteSiteID(path)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if spec.pagination == paginationOffset {
		url.WithQueryParam(limitParam, strconv.Itoa(pageSize(params)))
		url.WithQueryParam(offsetParam, "0")
	}

	// Comments are the only list that can be sorted server-side. Newest-first
	// by lastUpdated lets the time filter stop paging early.
	if spec.incrementalKey != "" {
		url.WithQueryParam(sortByParam, spec.incrementalKey)
		url.WithQueryParam(sortOrderParam, sortDescending)
	}

	return url, nil
}

// substituteSiteID fills the {site_id} placeholder with the connection's site.
// Paths without the placeholder (sites) pass through unchanged.
func (c *Connector) substituteSiteID(path string) (string, error) {
	if !strings.Contains(path, sitePlaceholder) {
		return path, nil
	}

	if c.SiteID == "" {
		return "", errMissingSiteID
	}

	return strings.ReplaceAll(path, sitePlaceholder, c.SiteID), nil
}

func pageSize(params common.ReadParams) int {
	return min(readhelper.PageSizeWithDefault(params, defaultPageSize), maxPageSize)
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	spec := readSpecs[params.ObjectName]

	requestURL, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)
	records := recordsFromPath(recordsKey)
	nextPage := nextPageFunc(spec, requestURL)
	marshal := readhelper.MakeMarshaledDataFuncWithId(nil, spec.idField)

	if spec.incrementalKey == "" {
		return common.ParseResult(resp, records, nextPage, marshal, params.Fields)
	}

	// Server-sorted newest first (comments): sieve the page by the timestamp
	// and stop paging at the first record older than Since. With no
	// Since/Until the filter passes every record through.
	return common.ParseResultFiltered(
		params,
		resp,
		records,
		readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder,
			readhelper.NewTimeBoundary(),
			spec.incrementalKey,
			timestampFormat,
			nextPage,
		),
		marshal,
		params.Fields,
	)
}

// recordsFromPath extracts the records array; a missing or null array yields
// no records, which ParseResult reports as an empty, finished page.
func recordsFromPath(recordsKey string) common.NodeRecordsFunc {
	return func(body *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(body).ArrayOptional(recordsKey)
	}
}

// nextPageFunc derives the next page from the response's pagination object:
// the next offset is offset+limit while that is below total. Objects without
// pagination (and responses that omit the object) end after one page.
func nextPageFunc(spec readSpec, requestURL *urlbuilder.URL) common.NextPageFunc {
	if spec.pagination != paginationOffset {
		return func(*ajson.Node) (string, error) { return "", nil }
	}

	return func(body *ajson.Node) (string, error) {
		nextOffset, ok, err := nextOffsetFromPagination(body)
		if err != nil || !ok {
			return "", err
		}

		nextURL := requestURL.Clone()
		nextURL.WithQueryParam(offsetParam, strconv.FormatInt(nextOffset, 10))

		return nextURL.String(), nil
	}
}

// nextOffsetFromPagination reads pagination{offset,limit,total} and reports
// whether another page exists. A missing or incomplete pagination object
// means the response was the whole collection.
func nextOffsetFromPagination(body *ajson.Node) (int64, bool, error) {
	pagination, err := jsonquery.New(body).ObjectOptional(paginationKey)
	if err != nil || pagination == nil {
		return 0, false, err
	}

	query := jsonquery.New(pagination)

	values := make([]*int64, 0, 3) //nolint:mnd // offset, limit, total

	for _, key := range []string{offsetParam, limitParam, "total"} {
		value, err := query.IntegerOptional(key)
		if err != nil {
			return 0, false, err
		}

		if value == nil {
			return 0, false, nil
		}

		values = append(values, value)
	}

	offset, limit, total := *values[0], *values[1], *values[2]

	nextOffset := offset + limit
	if nextOffset >= total {
		return 0, false, nil
	}

	return nextOffset, true, nil
}
