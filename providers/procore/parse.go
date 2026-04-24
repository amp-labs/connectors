package procore

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// nextPageFromLink returns a NextPageFunc that extracts the next page number
// from a Procore Link header via `Link: <...page=N>; rel="next"`.
// The returned token is the bare page number.
// Example Link header: `<https://api.procore.com/rest/v1.0/companies/12345/objects?page=2&per_page=100>; rel="next",
//
//	<https://api.procore.com/rest/v1.0/companies/12345/objects?page=50&per_page=100>; rel="last"`
//
// Ref: https://developers.procore.com/reference/rest/docs/pagination
func nextPageFromLink(linkHeader string) common.NextPageFunc {
	next := nextPageNumber(linkHeader)

	return func(*ajson.Node) (string, error) {
		return next, nil
	}
}

func nextPageNumber(linkHeader string) string {
	// If there is no Link header, we assume there are no more pages to fetch and return an empty token.
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

		parsed, err := url.Parse(part[start+1 : end])
		if err != nil {
			continue
		}

		return parsed.Query().Get("page")
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
