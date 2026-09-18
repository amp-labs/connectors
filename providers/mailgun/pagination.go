package mailgun

import (
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// nextPageFunc returns the pagination extractor for the object's style.
func (c *Connector) nextPageFunc(
	spec objectSpec, params common.ReadParams, reqURL *neturl.URL, recordsKey string,
) func(*ajson.Node) (string, error) {
	switch spec.pagination {
	case paginationCursor:
		return c.cursorNextPage
	case paginationOffset:
		return c.offsetNextPage(reqURL, recordsKey)
	case paginationLogsToken:
		return logsNextPage
	case paginationBodyOffset:
		return bodyOffsetNextPage(params, recordsKey)
	case paginationNone:
		fallthrough
	default:
		return noNextPage
	}
}

// bodyOffsetNextPage advances the body-carried skip (analytics/tags) by the
// returned record count; a short page ends the read. NextPage is the next skip
// value, which buildTagsRequest echoes into the following request body.
func bodyOffsetNextPage(params common.ReadParams, recordsKey string) func(*ajson.Node) (string, error) {
	return func(body *ajson.Node) (string, error) {
		records, err := common.ExtractOptionalRecordsFromPath(recordsKey)(body)
		if err != nil {
			return "", err
		}

		limit := mustPageSize(params)
		if len(records) < limit {
			return "", nil
		}

		return strconv.Itoa(nextPageInt(params.NextPage) + len(records)), nil
	}
}

// nextPageInt interprets NextPage as a skip offset; empty or invalid means 0.
func nextPageInt(token common.NextPageToken) int {
	value, err := strconv.Atoi(token.String())
	if err != nil || value < 0 {
		return 0
	}

	return value
}

// cursorNextPage reads paging.next and resolves it against the base URL when the
// provider returns a relative path (some v1 endpoints do). An empty items page
// is what actually stops pagination — handled by common.ParseResult — so we can
// unconditionally hand back paging.next here.
func (c *Connector) cursorNextPage(body *ajson.Node) (string, error) {
	next, err := pagingString(body, "paging", "next")
	if err != nil {
		return "", err
	}

	if next == "" {
		// Some v1 endpoints spell the paging keys capitalized — the OpenAPI spec
		// declares GET /v1/dynamic_pools/domains with paging {Next, Previous,
		// First, Last} — so fall back before concluding there is no next page.
		next, err = pagingString(body, "paging", "Next")
		if err != nil || next == "" {
			return "", err
		}
	}

	return c.absoluteURL(next), nil
}

// logsNextPage reads the raw pagination.next token from the logs response.
func logsNextPage(body *ajson.Node) (string, error) {
	return pagingString(body, "pagination", "next")
}

// pagingString reads container.key as a string, returning "" when either the
// container or the key is absent.
func pagingString(body *ajson.Node, container, key string) (string, error) {
	node, err := jsonquery.New(body).ObjectOptional(container)
	if err != nil || node == nil {
		return "", err
	}

	value, err := jsonquery.New(node).StringOptional(key)
	if err != nil || value == nil {
		return "", err
	}

	return *value, nil
}

// offsetNextPage advances skip by the returned record count until a short page
// signals the end.
func (c *Connector) offsetNextPage(reqURL *neturl.URL, recordsKey string) func(*ajson.Node) (string, error) {
	return func(body *ajson.Node) (string, error) {
		if reqURL == nil {
			return "", nil
		}

		records, err := common.ExtractOptionalRecordsFromPath(recordsKey)(body)
		if err != nil {
			return "", err
		}

		limit := queryParamInt(reqURL, limitParam, defaultPageSize)
		if len(records) < limit {
			return "", nil
		}

		next, err := urlbuilder.New(reqURL.String())
		if err != nil {
			return "", err
		}

		skip := queryParamInt(reqURL, skipParam, 0)
		next.WithQueryParam(skipParam, strconv.Itoa(skip+len(records)))

		return next.String(), nil
	}
}

// absoluteURL prefixes relative next-page paths with the connector base URL.
func (c *Connector) absoluteURL(next string) string {
	if strings.HasPrefix(next, "http://") || strings.HasPrefix(next, "https://") {
		return next
	}

	base := strings.TrimRight(c.ProviderInfo().BaseURL, "/")

	if !strings.HasPrefix(next, "/") {
		next = "/" + next
	}

	return base + next
}

func queryParamInt(u *neturl.URL, key string, defaultValue int) int {
	v, err := strconv.Atoi(u.Query().Get(key))
	if err != nil || v < 0 {
		return defaultValue
	}

	return v
}

// noNextPage is a pagination extractor for single-page fetches.
func noNextPage(*ajson.Node) (string, error) { return "", nil }
