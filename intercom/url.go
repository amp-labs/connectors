package intercom

import "github.com/amp-labs/connectors/common/urlbuilder"

// Intercom pagination cursor sometimes ends with `=`.
var intercomQueryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%3D": "=",
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(intercomQueryEncodingExceptions)

	return link, nil
}
