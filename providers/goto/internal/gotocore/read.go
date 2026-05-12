package gotocore

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	readPageSize       = "200"
	queryParamPage     = "page"
	queryParamPageSize = "pageSize"
	queryParamOffset   = "offset"
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	cfg, ok := objectRegistry[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: object %s is not registered for read",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	//nolint:exhaustive
	switch cfg.service {
	case serviceSCIM, serviceRemoteSupport, serviceCorporate, serviceMeetings:
		return a.buildUnpaginatedReadRequest(ctx, params)
	case serviceAdmin:
		return a.buildAdminReadRequest(ctx, params)
	default:
		return a.buildPagedReadRequest(ctx, params)
	}
}

// buildUnpaginatedReadRequest fetches the endpoint once with no pagination
// params. Used for services that don't support pagination at all
// (SCIM, Remote Support, Corporate).
func (a *Adapter) buildUnpaginatedReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	applyTimeFilter(url, params.ObjectName, params.Since, params.Until)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// buildAdminReadRequest paginates with `offset` + `pageSize`. The next-page
// token is the offset of the next record to fetch.
func (a *Adapter) buildAdminReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	applyTimeFilter(url, params.ObjectName, params.Since, params.Until)

	url.WithQueryParam(queryParamPageSize, readhelper.PageSizeWithDefaultStr(params, readPageSize))

	if params.NextPage != "" {
		url.WithQueryParam(queryParamOffset, params.NextPage.String())
	} else {
		url.WithQueryParam(queryParamOffset, "0")
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// buildPagedReadRequest is the default GoTo pagination: `size` + `page`.
// The next-page token is the next page number (0-indexed).
func (a *Adapter) buildPagedReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := a.buildObjectBaseURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(queryParamSize, readhelper.PageSizeWithDefaultStr(params, readPageSize))

	if params.NextPage != "" {
		url.WithQueryParam(queryParamPage, params.NextPage.String())
	} else {
		url.WithQueryParam(queryParamPage, "0")
	}

	applyTimeFilter(url, params.ObjectName, params.Since, params.Until)

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
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

// recordsExtractor returns a function that pulls the record array out of a
// GoTo response according to the service's envelope.
func recordsExtractor(service objectService, objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// Some objects ("historicalMeetings", "upcomingMeetings") are returned as a bare array at the root,
		// This handles that case.
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
		case serviceRemoteSupport:
			return readNodeArray(node, objectName)
			// Every other service wraps the records in an `_embedded` object, but the
			// key for the records array varies, so we pass it as a parameter to the
			// default reader.
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
// for the given service. SCIM, Corporate, and Remote Support don't paginate.
// Admin uses an offset+pageSize scheme. Everything else uses the default
// `page` envelope (number/totalPages).
func nextPageExtractor(service objectService) common.NextPageFunc {
	switch service { //nolint:exhaustive // page envelope is the default.
	case serviceSCIM, serviceCorporate, serviceRemoteSupport:
		return func(*ajson.Node) (string, error) { return "", nil }
	case serviceAdmin:
		return adminNextPage
	default:
		return webinarNextPage
	}
}

// webinarNextPage returns the next page number, or "" when there are no
// more pages. Page numbers are 0-indexed.
// if page object or its number/totalPages fields are missing,.
func webinarNextPage(node *ajson.Node) (string, error) {
	page := jsonquery.New(node, "page")

	// missing page object is normal.
	// some objects (e.g. userSubscriptions,webhooks) don't return a page object at all.
	// if the page object is missing or malformed, we assume there are no more pages.
	if page == nil {
		return "", nil //nolint:nilerr
	}

	currPage, err := page.IntegerRequired("number")
	if err != nil {
		return "", nil //nolint:nilerr
	}

	totalPages, err := page.IntegerRequired("totalPages")
	if err != nil {
		return "", nil //nolint:nilerr
	}

	next := currPage + 1
	if next >= totalPages {
		return "", nil
	}

	return strconv.FormatInt(next, 10), nil
}

// adminNextPage returns the next offset, or "" when there are no more
// records. The response reports `toIndex` (the last index returned, 0-indexed)
// and `total` (the total record count).
func adminNextPage(node *ajson.Node) (string, error) {
	query := jsonquery.New(node)

	toIndex, err := query.IntegerRequired("toIndex")
	if err != nil {
		return "", err //nolint:nilerr
	}

	total, err := query.IntegerRequired("total")
	if err != nil {
		return "", err //nolint:nilerr
	}

	next := toIndex + 1
	if next >= total {
		return "", nil
	}

	return strconv.FormatInt(next, 10), nil
}
