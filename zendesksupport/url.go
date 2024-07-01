package zendesksupport

import "github.com/amp-labs/connectors/common/urlbuilder"

// page size can be enclosed with square brackets.
var queryEncodingExceptions = map[string]string{} //nolint:gochecknoglobals

func constructURL(base string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base)
	if err != nil {
		return nil, err
	}

	link.AddEncodingExceptions(queryEncodingExceptions)

	return link, nil
}
