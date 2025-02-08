package chilipiper

import "github.com/amp-labs/connectors/common/urlbuilder"

const restAPIVersionPrefix = "api/fire-edge/v1/org"

func (conn *Connector) buildReadURL(object string) (*urlbuilder.URL, error) {
	return urlbuilder.New(conn.BaseURL, restAPIVersionPrefix, object)
}
