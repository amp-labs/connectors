package kit

import "github.com/amp-labs/connectors/common/urlbuilder"

// Intercom pagination cursor sometimes ends with `=`.
var intercomQueryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%3D": "=",
}

func constructURL(url *urlbuilder.URL, err error) (*urlbuilder.URL, error) {
	if err != nil {
		return nil, err
	}

	url.AddEncodingExceptions(intercomQueryEncodingExceptions)

	return url, nil
}
