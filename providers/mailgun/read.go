package mailgun

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 100

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	params = withIncrementalField(params)

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, strconv.Itoa(defaultPageSize))

	switch paginationForObject(params.ObjectName) {
	case readPaginationPagingNext, readPaginationLimitOnly:
		if params.ObjectName == "dynamic_pools/history" {
			endpointURL.WithQueryParam("Limit", pageSize)
		} else {
			endpointURL.WithQueryParam("limit", pageSize)
		}
	case readPaginationTotalCountSkip, readPaginationTotalSkip:
		endpointURL.WithQueryParam("limit", pageSize)
		endpointURL.WithQueryParam("skip", "0")
	case readPaginationNone:
	}

	applyNativeTimeFilters(endpointURL, params.ObjectName, params)

	return http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	return common.ParseResultFiltered(
		params,
		response,
		common.MakeRecordsFunc(responseFieldName),
		makeIncrementalFilterFunc(params, makeNextRecordsURL(params.ObjectName, request.URL)),
		common.MakeMarshaledDataFunc(nil),
		params.Fields,
	)
}

func makeNextRecordsURL(objectName string, requestURL *url.URL) common.NextPageFunc {
	strategy := paginationForObject(objectName)

	return func(node *ajson.Node) (string, error) {
		if requestURL == nil {
			return "", nil
		}

		switch strategy {
		case readPaginationPagingNext:
			return nextFromPaging(node, requestURL)
		case readPaginationTotalCountSkip:
			return nextFromSkip(node, requestURL, "total_count")
		case readPaginationTotalSkip:
			return nextFromSkip(node, requestURL, "total")
		case readPaginationLimitOnly, readPaginationNone:
			return "", nil
		default:
			return "", nil
		}
	}
}

func nextFromPaging(node *ajson.Node, requestURL *url.URL) (string, error) {
	items, err := jsonquery.New(node).ArrayOptional("items")
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "", nil
	}

	paging, err := jsonquery.New(node).ObjectOptional("paging")
	if err != nil || paging == nil {
		return "", nil //nolint:nilerr
	}

	// Most Mailgun list endpoints use lowercase paging.next.
	// Dynamic IP Pool endpoints document capital paging.Next.
	next := pagingNextLink(paging)
	if next == "" {
		return "", nil
	}

	parsed, err := url.Parse(next)
	if err != nil || (!strings.HasPrefix(next, "http") && !strings.HasPrefix(next, "/")) {
		return "", nil //nolint:nilerr
	}

	parsed.Scheme = requestURL.Scheme
	parsed.Host = requestURL.Host

	return parsed.String(), nil
}

func pagingNextLink(paging *ajson.Node) string {
	for _, key := range []string{"next", "Next"} {
		next, err := jsonquery.New(paging).StringOptional(key)
		if err != nil || next == nil || *next == "" {
			continue
		}

		return *next
	}

	return ""
}

func nextFromSkip(node *ajson.Node, requestURL *url.URL, totalField string) (string, error) {
	total, err := jsonquery.New(node).IntegerOptional(totalField)
	if err != nil || total == nil {
		return "", nil //nolint:nilerr
	}

	pageSize := defaultPageSize

	for _, key := range []string{"limit", "Limit"} {
		if limitStr := requestURL.Query().Get(key); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				pageSize = limit

				break
			}
		}
	}

	currentSkip := 0
	if skipStr := requestURL.Query().Get("skip"); skipStr != "" {
		currentSkip, err = strconv.Atoi(skipStr)
		if err != nil {
			return "", err
		}
	}

	if currentSkip+pageSize >= int(*total) {
		return "", nil
	}

	nextURL := *requestURL
	query := nextURL.Query()
	query.Set("skip", strconv.Itoa(currentSkip+pageSize))
	query.Set("limit", strconv.Itoa(pageSize))
	nextURL.RawQuery = query.Encode()

	return nextURL.String(), nil
}
