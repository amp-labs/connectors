package chilipiper

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIVersionPrefix = "api/fire-edge/v1/org"

func (conn *Connector) buildURL(objectName string, pageSize string) (string, error) {
	path, err := supportsRead(objectName)
	if err != nil {
		return "", err
	}

	url, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, path)
	if err != nil {
		return "", err
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return url.String(), nil
}

func (conn *Connector) buildWriteURL(object string) (*urlbuilder.URL, error) {
	path, err := supportsWrite(object)
	if err != nil {
		return nil, err
	}

	writeURL, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, path)
	if err != nil {
		return nil, err
	}

	return writeURL, nil
}
