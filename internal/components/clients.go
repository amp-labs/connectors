package components

import (
	"github.com/amp-labs/connectors/common"
)

// AuthClients implements the HTTP client accessors required by connectors.Connector.
//
// It builds an authenticated HTTP client and a JSON HTTP client backed by the
// supplied common.AuthenticatedHTTPClient.
// AuthClients is intended to be embedded in connector implementations.
// It is not intended to be used as a standalone connector.
//
// Keep this type limited to the supported HTTP clients. Do not add clients for
// other media types or protocols, such as XML, HTML, or SQL.
type AuthClients struct {
	http *common.HTTPClient
	json *common.JSONHTTPClient
}

// NewAuthClients creates the authenticated HTTP clients whose accessor
// methods satisfy part of the connectors.Connector interface when AuthClients
// is embedded in a connector implementation.
//
// The returned clients do not configure a base URL or the legacy error and
// response-handling fields. Those fields exist only for backward compatibility
// and should not be used by new code.
//
// Connector operations should handle errors through the appropriate
// operations.HTTPHandlers instead.
func NewAuthClients(
	client common.AuthenticatedHTTPClient,
) (*AuthClients, error) {
	httpClient := &common.HTTPClient{
		Client:            client,
		Base:              "",  // not set -- deprecated
		ErrorHandler:      nil, // not set -- deprecated
		ResponseHandler:   nil, // not set -- deprecated
		ShouldHandleError: nil, // not set -- deprecated
	}

	jsonClient := &common.JSONHTTPClient{
		HTTPClient:         httpClient,
		ErrorPostProcessor: common.ErrorPostProcessor{}, // not set -- deprecated
	}

	return &AuthClients{
		http: httpClient,
		json: jsonClient,
	}, nil
}

func (c *AuthClients) JSONHTTPClient() *common.JSONHTTPClient {
	return c.json
}

func (c *AuthClients) HTTPClient() *common.HTTPClient {
	return c.http
}
