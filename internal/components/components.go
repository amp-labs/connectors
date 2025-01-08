package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
)

type Setup[T any] func(component *ConnectorComponent) (*T, error)

type ConnectorComponent struct {
	// These make up the base behavior of the connector, defining how a connector talks to an API, and what operations
	// are allowed on the connector.
	ClientComponent
	ProviderEndpointSupport
}

func Initialize[T any](
	provider providers.Provider,
	params common.Parameters,
	setup Setup[T],
	opts ...Option,
) (conn *T, err error) {
	defer goutils.PanicRecovery(func(cause error) {
		err = cause
		conn = nil
	})

	clientComponent, err := NewClientComponent(provider, params)
	if err != nil {
		return nil, err
	}

	connectorComponent := &ConnectorComponent{ClientComponent: *clientComponent}

	// Apply options *before* the setup, so that setups can override if needed
	for _, opt := range opts {
		opt(connectorComponent)
	}

	conn, err = setup(connectorComponent)
	if err != nil {
		return nil, err
	}

	if err := common.ValidateParameters(conn, params); err != nil {
		return nil, err
	}

	return conn, nil
}
