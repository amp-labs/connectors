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

const (
	readPageSize   = "200"
	queryParamPage = "page"
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	cfg, ok := objectRegistry[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: object %s is not registered for read",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	// SCIM endpoints (users, groups) don't support pagination on GoTo —
	// fetch once and return everything in a single page.
	if cfg.service == serviceSCIM {
		url, err := a.buildObjectBaseURL(params.ObjectName)
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	// Webinar responses don't include a next-page URL, so we rebuild the URL
	// and pass the next page number as `?page=N`.
	if cfg.service == serviceWebinar {
		url, err := a.buildObjectBaseURL(params.ObjectName)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam(queryParamSize, readPageSize)
		if params.NextPage != "" {
			url.WithQueryParam(queryParamPage, params.NextPage.String())
		} else {
			url.WithQueryParam(queryParamPage, "0")
		}

		applyTimeFilter(url, params.ObjectName, params.Since, params.Until)

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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

	url.WithQueryParam(queryParamSize, readPageSize)

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
		common.GetMarshaledData,
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
		case serviceAssist:
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
// for the given service. SCIM endpoints don't paginate. Webinar uses a `page`
// envelope (number/totalPages) and emits the next page number as the token.
// The rest expose a HAL `_links.next.href` URL that the next request can use
// directly.
func nextPageExtractor(service objectService) common.NextPageFunc {
	switch service { //nolint:exhaustive // HAL is the default.
	case serviceSCIM, serviceCorporate, serviceAssist:
		return func(*ajson.Node) (string, error) { return "", nil }
	case serviceWebinar:
		return webinarNextPage

	default:
		return halNextPage
	}
}

func halNextPage(node *ajson.Node) (string, error) {
	href, err := jsonquery.New(node, "_links", "next").StringOptional("href")
	if err != nil || href == nil {
		return "", nil //nolint:nilerr // missing next link is normal end-of-pagination.
	}

	return *href, nil
}

// webinarNextPage returns the next page number, or "" when there are no
// more pages. Page numbers are 0-indexed.
// if page object or its number/totalPages fields are missing,
func webinarNextPage(node *ajson.Node) (string, error) {
	page := jsonquery.New(node, "page")

	// missing page object is normal.
	// some objects (e.g. userSubscriptions,webhooks) don't return a page object at all.
	// if the page object is missing or malformed, we assume there are no more pages.
	if page == nil {
		return "", nil //nolint:nilerr
	}

	CurrPage, err := page.IntegerRequired("number")
	if err != nil {
		return "", nil //nolint:nilerr
	}

	totalPages, err := page.IntegerRequired("totalPages")
	if err != nil {
		return "", nil //nolint:nilerr
	}

	next := CurrPage + 1
	if next >= totalPages {
		return "", nil
	}

	return strconv.FormatInt(next, 10), nil
}
