package smartlead

import "github.com/amp-labs/connectors/common/urlbuilder"

var queryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	// none
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(queryEncodingExceptions)

	return link, nil
}
