package helpscout

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const restAPIVersion = "v2"

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func (conn *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(conn.BaseURL, restAPIVersion, objectName)
}
