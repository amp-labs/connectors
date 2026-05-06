package procore

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// nextPageFromLink returns a NextPageFunc that extracts the full next-page URL
// from a Procore Link header via `Link: <...>; rel="next"`.
//
// We forward the entire URL (not just the page number) so that filter params
// captured on the first request — most importantly `filters[updated_at]` — are
// carried across pages. This keeps the time window stable and avoids
// duplicates/misses that would occur if the connector recomputed the window
// (e.g. defaulting `until` to time.Now()) for each page.
//
// Example Link header: `<https://api.procore.com/rest/v1.0/companies/12345/objects?page=2&per_page=100>; rel="next",
//
//	<https://api.procore.com/rest/v1.0/companies/12345/objects?page=50&per_page=100>; rel="last"`
//
// Ref: https://developers.procore.com/reference/rest/docs/pagination
func nextPageFromLink(linkHeader string) common.NextPageFunc {
	next := nextPageURL(linkHeader)

	return func(*ajson.Node) (string, error) {
		return next, nil
	}
}

// nextPageURL returns the absolute URL for the rel="next" entry in a Procore
// Link header, or an empty string when there is no next page.
func nextPageURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	for part := range strings.SplitSeq(linkHeader, ",") {
		if !strings.Contains(part, `rel="next"`) {
			continue
		}

		start := strings.Index(part, "<")

		end := strings.Index(part, ">")
		if start < 0 || end <= start {
			continue
		}

		raw := part[start+1 : end]

		if _, err := url.Parse(raw); err != nil {
			continue
		}

		return raw
	}

	return ""
}

// extractRecords returns the list of records from a Procore response.
// When the registry declares a records key, unwrap that key; otherwise
// treat the response body as the array itself.
func extractRecords(response *common.JSONHTTPResponse, objectName string) ([]any, error) {
	responseKey := objectRegistry[objectName].recordsKey

	if responseKey != "" {
		obj, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil || obj == nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		data, ok := (*obj)[responseKey].([]any)
		if !ok {
			return nil, fmt.Errorf("%w: response object missing array under \"%s\" key",
				common.ErrMissingExpectedValues, responseKey)
		}

		return data, nil
	}

	arr, err := common.UnmarshalJSON[[]any](response)
	if err == nil {
		return *arr, nil
	}

	return nil, fmt.Errorf("response body is not in an expected format: %w", common.ErrMissingExpectedValues)
}
