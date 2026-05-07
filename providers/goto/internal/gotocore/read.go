package gotocore

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const readPageSize = "100"

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	cfg, ok := objectRegistry[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: object %s is not registered for read",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	// SCIM paginates by startIndex; for that service we encode the next
	// startIndex as the opaque token rather than a full URL.
	if cfg.service == serviceSCIM && params.NextPage != "" {
		return a.buildSCIMReadRequest(ctx, params)
	}

	// Every other GoTo service returns a full HAL `_links.next.href`,
	// which we can use verbatim as the next request URL.
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if cfg.service == serviceSCIM {
		url.WithQueryParam("count", readPageSize)
	} else {
		url.WithQueryParam(queryParamSize, readPageSize)
	}

	applyTimeFilter(url, params.ObjectName, params.Since, params.Until)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) buildSCIMReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("count", readPageSize)
	url.WithQueryParam("startIndex", params.NextPage.String())

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	cfg, ok := objectRegistry[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: object %s is not registered for read",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	return common.ParseResult(
		resp,
		recordsExtractor(cfg.service, params.ObjectName),
		nextPageExtractor(cfg.service),
		common.GetMarshaledData,
		params.Fields,
	)
}

// recordsExtractor returns a function that pulls the record array out of a
// GoTo response according to the service's envelope.
func recordsExtractor(service objectService, objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		if node.IsArray() {
			arr, err := node.GetArray()
			if err != nil {
				return nil, err
			}

			return jsonquery.Convertor.ArrayToMap(arr)
		}

		switch service { //nolint:exhaustive // _embedded.<obj> shape is the default.
		case serviceSCIM:
			return readNodeArray(node, "resources")
		case serviceAdmin:
			return readNodeArray(node, "results")
		case serviceAssist:
			return readNodeArray(node, objectName)
		default:
			return readNodeArray(node, objectName, "_embedded")
		}
	}
}

func readNodeArray(node *ajson.Node, jsonPath string, nestedPath ...string) ([]map[string]any, error) {
	arr, err := jsonquery.New(node, nestedPath...).ArrayOptional(jsonPath)
	if err != nil {
		return nil, err
	}

	if arr == nil {
		return nil, nil
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

// nextPageExtractor returns a NextPageFunc that resolves the next-page token
// for the given service. SCIM uses startIndex+itemsPerPage; the rest expose
// a HAL `_links.next.href` URL that the next request can use directly.
func nextPageExtractor(service objectService) common.NextPageFunc {
	if service == serviceSCIM {
		return scimNextPage
	}

	return halNextPage
}

func halNextPage(node *ajson.Node) (string, error) {
	href, err := jsonquery.New(node, "_links", "next").StringOptional("href")
	if err != nil || href == nil {
		return "", nil //nolint:nilerr // missing next link is normal end-of-pagination.
	}

	return *href, nil
}

func scimNextPage(node *ajson.Node) (string, error) {
	query := jsonquery.New(node)

	startIndex, err := query.IntegerOptional("startIndex")
	if err != nil || startIndex == nil {
		return "", nil //nolint:nilerr
	}

	itemsPerPage, err := query.IntegerOptional("itemsPerPage")
	if err != nil || itemsPerPage == nil {
		return "", nil //nolint:nilerr
	}

	totalResults, err := query.IntegerOptional("totalResults")
	if err != nil || totalResults == nil {
		return "", nil //nolint:nilerr
	}

	next := *startIndex + *itemsPerPage
	if next > *totalResults {
		return "", nil
	}

	return strconv.FormatInt(next, 10), nil
}
