package intercom

import "github.com/amp-labs/connectors/common/urlbuilder"

// Intercom pagination cursor sometimes ends with `=`.
var intercomQueryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%3D": "=",
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	url.AddEncodingExceptions(intercomQueryEncodingExceptions)

	return url, nil
}
