package dynamicscrm

import "github.com/amp-labs/connectors/common/urlbuilder"

// Microsoft uses special symbology when making queries.
var queryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%40": "@",
	"%24": "$",
	"%2C": ",",
}

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(queryEncodingExceptions)

	return link, nil
}
