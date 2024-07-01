package gong

import "github.com/amp-labs/connectors/common/urlbuilder"

var queryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	// none
}

func constructURL(base string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(queryEncodingExceptions)

	return link, nil
}
