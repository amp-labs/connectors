package chilipiper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIVersionPrefix = "api/fire-edge/v1/org"

func (conn *Connector) buildURL(objectName string, pageSize string) (string, error) {
	if !supportedReadObjects.Has(objectName) {
		return "", common.ErrObjectNotSupported
	}

	url, err := urlbuilder.New(conn.ProviderInfo().BaseURL, restAPIVersionPrefix, objectName)
	if err != nil {
		return "", err
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return url.String(), nil
}

func (conn *Connector) buildWriteURL(object string) (*urlbuilder.URL, error) {
	if !supportedWriteObjects.Has(object) {
		return nil, common.ErrObjectNotSupported
	}

	writeURL, err := urlbuilder.New(conn.ProviderInfo().BaseURL, restAPIVersionPrefix, object)
	if err != nil {
		return nil, err
	}

	return writeURL, nil
}
