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
