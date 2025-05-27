package mockutils

import "github.com/amp-labs/connectors/common/urlbuilder"

func ReplaceURLOrigin(baseURL, mockServerOrigin string) (alteredBaseURL string) {
	url, err := urlbuilder.New(baseURL)
	if err != nil {
		// Provider URL must follow valid URL format.
		return "impossible"
	}

	return mockServerOrigin + url.Path()
}
