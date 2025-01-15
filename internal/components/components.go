package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
)

// ConnectorConstructor is a function that creates a new instance of a connector T, given a ConnectorComponent.
type ConnectorConstructor[T any] func(component *ConnectorComponent) (*T, error)

type ConnectorComponent struct {
	// These make up the base behavior of the connector, defining how a connector talks to an API, and what operations
	// are allowed on the connector.
	ClientComponent
	ProviderEndpointSupport
}

// InitializeConnector initializes a connector T with the given provider, parameters, and constructor.
func InitializeConnector[T any](
	provider providers.Provider,
	params common.Parameters,
	constructor ConnectorConstructor[T],
	opts ...Option,
) (conn *T, err error) {
	defer goutils.PanicRecovery(func(cause error) {
		err = cause
		conn = nil
	})

	connectorComponent, err := initializeConnectorComponent(provider, params, opts...)
	if err != nil {
		return nil, err
	}

	conn, err = constructor(connectorComponent)
	if err != nil {
		return nil, err
	}

	if err := common.ValidateParameters(conn, params); err != nil {
		return nil, err
	}

	return conn, nil
}

func initializeConnectorComponent(
	provider providers.Provider,
	params common.Parameters,
	opts ...Option,
) (*ConnectorComponent, error) {
	clientComponent, err := NewClientComponent(provider, params)
	if err != nil {
		return nil, err
	}

	connectorComponent := &ConnectorComponent{ClientComponent: *clientComponent}

	// Apply options *before* the constructor, so that constructors can override if needed
	for _, opt := range opts {
		opt(connectorComponent)
	}

	return connectorComponent, nil
}
