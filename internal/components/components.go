package components

import (
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
)

type Setup[T any] func(component *ConnectorComponent) (*T, error)

type ConnectorComponent struct {
	*ClientComponent
	MetadataStrategy
}

func Initialize[T any](
	provider providers.Provider,
	params connector.Parameters,
	setup Setup[T],
) (conn *T, err error) {
	defer goutils.PanicRecovery(func(cause error) {
		err = cause
		conn = nil
	})

	clientComponent, err := NewClientComponent(provider, params)
	if err != nil {
		return nil, err
	}

	connectorComponent := &ConnectorComponent{ClientComponent: clientComponent}
	return setup(connectorComponent)
}
