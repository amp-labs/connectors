package monaco

import (
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

// buildURL joins path segments onto the base URL, preserving a trailing slash
// on the last segment when it has one.
//
// Monaco's routes are slash-exact and inconsistent about it. Probing the live
// API: /v1/contacts/, /v1/tags/, /v1/tasks/, /v1/campaigns/, /v1/opportunities/
// and /v1/accounts/ are served directly while their unslashed forms answer 307,
// and /v1/audiences and /v1/sequence-templates are the exact reverse. urlbuilder
// normalizes trailing slashes away unconditionally, so without this the
// connector would lean on redirects for roughly half its endpoints.
func buildURL(baseURL string, parts ...string) (string, error) {
	endpointURL, err := urlbuilder.New(baseURL, parts...)
	if err != nil {
		return "", err
	}

	result := endpointURL.String()

	if wantsTrailingSlash(parts) && !strings.HasSuffix(result, "/") {
		result += "/"
	}

	return result, nil
}

func wantsTrailingSlash(parts []string) bool {
	for _, part := range slices.Backward(parts) {
		if part == "" {
			continue
		}

		return strings.HasSuffix(part, "/")
	}

	return false
}
