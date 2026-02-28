package components

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
)

var ErrSupportNotConfigured = errors.New("support not configured")

// ConnectorConstructor is a function that constructs a connector from a connector and an endpoint supportregistry.
// TODO: Convert this to a type alias for easier usage when we go to go1.24: https://go.dev/doc/go1.24#language
type ConnectorConstructor[T any] func(*Connector) (*T, error)

// ConnectorConstructorWithParams is a function type for creating a connector instance
// by building on a base Connector.
//
// The base Connector represents a provider from the catalog (providers.ProviderInfo)
// and includes any values that are inferred or bootstrapped. Specific connectors
// can embed this base to extend its behavior.
//
// It receives:
//   - params: connector initialization parameters (ConnectorParams)
//   - base: the initialized Connector built from the provider info
//
// Returns the typed connector (*T) or an error.
//
// Example:
//
//	var myConstructor ConnectorConstructorWithParams[MyConnectorType] =
//	    func(params common.ConnectorParams, base *Connector) (*MyConnectorType, error) {
//	        // embed base and initialize MyConnectorType
//	    }
type ConnectorConstructorWithParams[T any] func(common.ConnectorParams, *Connector) (*T, error)

// Connector provides a reusable base for API connectors, embedding Transport
// and explicitly defining core methods (JSONHTTPClient, HTTPClient, Provider, String)
// to avoid ambiguity when combined with interfaces that embed fmt.Stringer.
type Connector struct {
	*Transport
}

// Initialize initializes a connector with the given provider and parameters
// by using Connector as a base type. It runs the constructor with the connector
// and returns the connector as the specified T type.
//
// Deprecated: use Init.
func Initialize[T any](
	provider providers.Provider,
	params common.ConnectorParams,
	constructor ConnectorConstructor[T],
) (conn *T, err error) {
	return Init(provider, params, func(_ common.ConnectorParams, connector *Connector) (*T, error) {
		return constructor(connector)
	})
}

// Init initializes a connector with the given provider and parameters
// by using Connector as a base type. It runs the constructor with the connector
// and returns the connector as the specified T type.
func Init[T any](
	provider providers.Provider,
	params common.ConnectorParams,
	constructor ConnectorConstructorWithParams[T],
) (conn *T, err error) {
	defer goutils.PanicRecovery(func(cause error) {
		err = cause
		conn = nil
	})

	// Default module is always the root module
	if params.Module == "" {
		params.Module = common.ModuleRoot
	}

	transport, err := NewTransport(provider, params)
	if err != nil {
		return nil, err
	}

	conn, err = constructor(params, &Connector{Transport: transport})
	if err != nil {
		return nil, err
	}

	// Validate the parameters for the connector
	if err := common.ValidateParameters(conn, params); err != nil {
		return nil, err
	}

	return conn, nil
}

// JSONHTTPClient returns the connector's JSON-capable HTTP client.
// Defined explicitly to expose the method on Connector for compile-time conflict detection with interfaces.
func (c Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Transport.JSONHTTPClient() // do not remove
}

// HTTPClient returns the underlying raw HTTP client.
// Defined explicitly to expose the method on Connector for compile-time conflict detection with interfaces.
func (c Connector) HTTPClient() *common.HTTPClient {
	return c.Transport.HTTPClient()
}

// Provider returns the provider associated with this connector.
// Defined explicitly to expose the method on Connector for compile-time conflict detection with interfaces.
func (c Connector) Provider() providers.Provider {
	return c.Transport.ProviderContext.Provider()
}

// String returns a human-readable identifier for this connector.
// Defined explicitly to expose the method on Connector for compile-time conflict detection with interfaces.
func (c Connector) String() string {
	return c.Transport.ProviderContext.String()
}
