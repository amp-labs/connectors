package microsoftdynamicscrm

import "github.com/amp-labs/connectors/common/urlbuilder"

// Microsoft uses special symbology when making queries.
var microsoftQueryEncodingExceptions = map[string]string{ //nolint:gochecknoglobals
	"%40": "@",
	"%24": "$",
	"%2C": ",",
}

func constructURL(base string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(microsoftQueryEncodingExceptions)

	return link, nil
}
