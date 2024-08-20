package marketo

import "github.com/amp-labs/connectors/common/urlbuilder"

var restAPIPrefix string = "rest" //nolint:gochecknoglobals

func constructURL(base string, path ...string) (*urlbuilder.URL, error) {
	link, err := urlbuilder.New(base, path...)
	if err != nil {
		return nil, err
	}

	return link, nil
}
