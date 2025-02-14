package front

import "github.com/amp-labs/connectors/common"

// JSONHTTPClient returns the underlying JSON HTTP client.
func (conn *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return conn.Client
}

func (conn *Connector) HTTPClient() *common.HTTPClient {
	return conn.Client.HTTPClient
}
