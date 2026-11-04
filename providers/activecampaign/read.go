package activecampaign

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/activecampaign/internal/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "api/3"

	// Endpoints return 20 records by default and accept at most 100.
	// https://developers.activecampaign.com/reference/pagination
	defaultPageSize = 100
	maxPageSize     = 100

	limitQuery     = "limit"
	offsetQuery    = "offset"
	idGreaterQuery = "id_greater"
	orderByIDQuery = "orders[id]"
	orderAscending = "ASC"
)

// Objects paginated by walking forward over record identifiers rather than by offset.
// ActiveCampaign warns that offset pagination slows down on accounts holding many records and
// recommends sorting by identifier ascending and requesting the next batch with id_greater.
// That parameter is currently only available on the contacts endpoint.
// https://developers.activecampaign.com/reference/pagination
var cursorPaginatedObjects = datautils.NewSet("contacts") // nolint:gochecknoglobals

// Objects that support provider-side incremental read via a time based query parameter.
// https://developers.activecampaign.com/reference/list-all-contacts
// https://developers.activecampaign.com/reference/list-all-deals
var incrementalReadParams = map[string]string{ // nolint:gochecknoglobals
	"contacts": "filters[updated_after]",
	"deals":    "filters[updated_after]",
}

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page URL is complete and can be used verbatim.
		return urlbuilder.New(params.NextPage.String())
	}

	// First page.
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(limitQuery, strconv.Itoa(pageSize(params)))

	if cursorPaginatedObjects.Has(params.ObjectName) {
		// Ascending identifier order is what lets id_greater walk the complete result set.
		url.WithQueryParam(orderByIDQuery, orderAscending)
	}

	if queryParam, ok := incrementalReadParams[params.ObjectName]; ok && !params.Since.IsZero() {
		url.WithQueryParam(queryParam, datautils.Time.FormatRFC3339inUTC(params.Since))
	}

	return url, nil
}

// pageSize honours the requested page size, defaulting to the largest batch the API allows.
// Values above the cap are rejected by the provider with 422, so they are clamped instead.
func pageSize(params common.ReadParams) int {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		return defaultPageSize
	}

	return params.PageSize
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	numRecords := 0
	lastRecordID := ""

	if body, ok := response.Body(); ok {
		records, err := jsonquery.New(body).ArrayOptional(params.ObjectName)
		if err != nil {
			return nil, err
		}

		numRecords = len(records)

		if numRecords != 0 {
			lastRecordID, err = jsonquery.New(records[numRecords-1]).TextWithDefault("id", "")
			if err != nil {
				return nil, err
			}
		}
	}

	url, err := urlbuilder.New(request.URL.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		// Records are located under the key matching the object name,
		// ex: {"contacts": [...], "meta": {"total": "163"}}.
		common.ExtractOptionalRecordsFromPath(params.ObjectName),
		makeNextRecordsURL(url, params.ObjectName, numRecords, lastRecordID),
		common.GetMarshaledData,
		params.Fields,
	)
}

// makeNextRecordsURL selects the pagination strategy the object supports.
func makeNextRecordsURL(
	previousURL *urlbuilder.URL,
	objectName string,
	numRecords int,
	lastRecordID string,
) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if numRecords == 0 {
			return "", nil
		}

		if cursorPaginatedObjects.Has(objectName) {
			return nextCursorURL(previousURL, numRecords, lastRecordID)
		}

		return nextOffsetURL(previousURL, node, numRecords)
	}
}

// nextCursorURL resumes after the highest identifier seen on the current page.
// Records arrive in ascending identifier order, so a partial page means the collection is exhausted.
func nextCursorURL(previousURL *urlbuilder.URL, numRecords int, lastRecordID string) (string, error) {
	if numRecords < requestedPageSize(previousURL) {
		return "", nil
	}

	if lastRecordID == "" {
		// The cursor cannot advance without an identifier; stop rather than repeat this page.
		return "", nil
	}

	previousURL.WithQueryParam(idGreaterQuery, lastRecordID)

	return previousURL.String(), nil
}

// nextOffsetURL advances the offset query parameter on the previous request URL.
// ActiveCampaign reports the total number of records as a string under meta.total
// but does not echo the offset back, ex: {"meta": {"total": "163"}}.
func nextOffsetURL(previousURL *urlbuilder.URL, node *ajson.Node, numRecords int) (string, error) {
	prevOffsetStr, found := previousURL.GetFirstQueryParam(offsetQuery)
	if !found {
		prevOffsetStr = "0"
	}

	prevOffset, err := strconv.Atoi(prevOffsetStr)
	if err != nil {
		return "", err
	}

	nextOffset := prevOffset + numRecords

	total, err := extractTotalCount(node)
	if err != nil {
		return "", err
	}

	if total != nil {
		if int64(nextOffset) >= *total {
			return "", nil
		}
	} else if numRecords < requestedPageSize(previousURL) {
		// Without a total count a short page means there is no more data.
		return "", nil
	}

	previousURL.WithQueryParam(offsetQuery, strconv.Itoa(nextOffset))

	return previousURL.String(), nil
}

// requestedPageSize reports the limit that produced the response currently being parsed.
func requestedPageSize(previousURL *urlbuilder.URL) int {
	sizeStr, found := previousURL.GetFirstQueryParam(limitQuery)
	if !found {
		return defaultPageSize
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return defaultPageSize
	}

	return size
}

func extractTotalCount(node *ajson.Node) (*int64, error) {
	meta, err := jsonquery.New(node).ObjectOptional("meta")
	if err != nil {
		return nil, err
	}

	if meta == nil {
		return nil, nil // nolint:nilnil
	}

	totalStr, err := jsonquery.New(meta).TextWithDefault("total", "")
	if err != nil {
		return nil, err
	}

	if totalStr == "" {
		return nil, nil // nolint:nilnil
	}

	total, err := strconv.ParseInt(totalStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &total, nil
}
