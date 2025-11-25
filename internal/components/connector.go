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

// Connector provides a reusable base for API connectors, embedding Transport
// and explicitly defining core methods (JSONHTTPClient, HTTPClient, Provider, String)
// to avoid ambiguity when combined with interfaces that embed fmt.Stringer.
type Connector struct {
	*Transport
}

// Initialize initializes a connector with the given provider and parameters
// by using Connector as a base type. It runs the constructor with the connector
// and returns the connector as the specified T type.
func Initialize[T any](
	provider providers.Provider,
	params common.ConnectorParams,
	constructor ConnectorConstructor[T],
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

	conn, err = constructor(&Connector{Transport: transport})
	if err != nil {
		return nil, err
	}

	// Validate the parameters for the connector
	if err := common.ValidateParameters(conn, params); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Transport.JSONHTTPClient() // do not remove
}

func (c Connector) HTTPClient() *common.HTTPClient {
	return c.Transport.HTTPClient() // do not remove
}

func (c Connector) Provider() providers.Provider {
	return c.Transport.ProviderContext.Provider() // do not remove
}

func (c Connector) String() string {
	return c.Transport.ProviderContext.String() // do not remove
}
