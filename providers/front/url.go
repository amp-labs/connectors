package front

import "github.com/amp-labs/connectors/common"

const (
	pageSizeKey         = "limit"
	paginationResultKey = "_pagination"
	nextURLKey          = "next"
	metadataPageSize    = "1"
	pageSize            = "100"
)

func (conn *Connector) buildURL(objectName string, pageSize string) (string, error) {
	if !supportsRead(objectName) {
		return "", common.ErrObjectNotSupported
	}

	url, err := conn.getBaseAPIURL(objectName)
	if err != nil {
		return "", err
	}

	url.WithQueryParam(pageSizeKey, pageSize)

	return url.String(), nil
}
