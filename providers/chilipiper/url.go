package chilipiper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersionPrefix = "api/fire-edge/v1/org"
	meetings             = "meetings/meetings"
)

func (conn *Connector) buildURL(objectName string, pageSize string) (*urlbuilder.URL, error) {
	if !supportedReadObjects.Has(objectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return url, nil
}

func (conn *Connector) buildWriteURL(object string) (*urlbuilder.URL, error) {
	if !supportedWriteObjects.Has(object) {
		return nil, common.ErrObjectNotSupported
	}

	writeURL, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, object)
	if err != nil {
		return nil, err
	}

	return writeURL, nil
}
