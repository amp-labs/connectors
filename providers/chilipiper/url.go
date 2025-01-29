package chilipiper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIVersionPrefix = "api/fire-edge/v1/org"

func (conn *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	path, err := supportsRead(config.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, path)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam(pageSizeKey, pageSize)

	return readURL, nil
}

func (conn *Connector) buildMetadataURL(object string) (*urlbuilder.URL, error) {
	path, err := supportsRead(object)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, path)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam(pageSizeKey, metadataPageSize)

	return readURL, nil
}
