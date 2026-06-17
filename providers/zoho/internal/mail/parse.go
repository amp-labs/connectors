package mail

import (
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// makeNextRecordsURL uses three pagination styles to build the URL for the next page, based on the response data:
// 1) nextPageFullURL: a full URL is returned at data.pagination.next (e.g. notes and links families).
// 2) nextPageRelativeURL: a path relative to the API root is returned at data.paging.nextPage (e.g. tasks).
// 3) nextPageOffset: no next-page URL is returned; we advance the offset ourselves (e.g. messages).
func (a *Adapter) makeNextRecordsURL(reqURL *urlbuilder.URL, obj objectDescriptor) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// If the object doesn't support pagination
		//  we return nil to end the read after the first page.
		if obj.pagination == nil {
			return "", nil
		}

		switch obj.pagination.style {
		case nextPageFullURL:
			// Full URL at data.pagination.next; empty when there are no more.
			return readNextPage(node, "pagination", "next"), nil
		case nextPageRelativeURL:
			// data.paging.nextPage is a path relative to the API root,
			// e.g. "tasks/me?from=2&limit=2" -> {baseURL}/api/tasks/me?...
			rel := readNextPage(node, "paging", "nextPage")
			if rel == "" {
				return "", nil
			}

			return strings.TrimRight(a.BaseURL, "/") + "/api/" + rel, nil
		case nextPageOffset:
			return a.offsetNextPage(reqURL, obj, node)
		default:
			return "", nil
		}
	}
}

// readNextPage reads the next-page string at data.<group>.<key>.
// when the response carries none (key missing, or data is not an object).
func readNextPage(node *ajson.Node, group, key string) string {
	value, err := jsonquery.New(node, "data", group).StrWithDefault(key, "")
	if err != nil {
		return ""
	}

	return value
}

// OffsetNextPage build the next page URL using request URL and response data.
func (a *Adapter) offsetNextPage(
	reqURL *urlbuilder.URL, obj objectDescriptor, node *ajson.Node,
) (string, error) {
	records, err := extractRecordsFromKeyPath(obj.recordsPath)(node)
	if err != nil {
		return "", err
	}

	limitStr, ok := reqURL.GetFirstQueryParam("limit") //nolint:varnamelen
	if !ok {
		return "", nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return "", err
	}

	// If the number of records returned is less than the page size, there are no more pages.
	// (Zoho Mail doesn't return total counts, so we can't know the total number of pages upfront.)
	if len(records) == 0 || len(records) < limit {
		return "", nil
	}

	offsetStr, ok := reqURL.GetFirstQueryParam(obj.pagination.offsetParam)
	if !ok {
		return "", nil
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return "", err
	}

	reqURL.WithQueryParam(obj.pagination.offsetParam, strconv.Itoa(offset+limit))

	return reqURL.String(), nil
}
