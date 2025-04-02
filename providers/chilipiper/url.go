package chilipiper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (conn *Connector) buildURL(objectName string, pageSize string) (string, error) {
	if !supportedReadObjects.Has(objectName) {
		return "", common.ErrObjectNotSupported
	}

	url, err := conn.ModuleClient.URL(objectName)
	if err != nil {
		return "", err
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return url.String(), nil
}

func (conn *Connector) buildWriteURL(objectName string) (*urlbuilder.URL, error) {
	if !supportedWriteObjects.Has(objectName) {
		return nil, common.ErrObjectNotSupported
	}

	writeURL, err := conn.ModuleClient.URL(objectName)
	if err != nil {
		return nil, err
	}

	return writeURL, nil
}
