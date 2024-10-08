package deep

import (
	"errors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

type ConnectorBuilder[C any, P paramsbuilder.ParamAssurance] struct {
	provider     providers.Provider
	errorHandler interpreter.ErrorHandler
	setup        func(conn *C)
}

func Connector[C any, P paramsbuilder.ParamAssurance](
	provider providers.Provider,
	errorHandler interpreter.ErrorHandler) *ConnectorBuilder[C, P] {
	return &ConnectorBuilder[C, P]{
		provider:     provider,
		errorHandler: errorHandler,
	}
}

func (b *ConnectorBuilder[C, P]) Setup(setup func(conn *C)) *ConnectorBuilder[C, P] {
	b.setup = setup

	return b
}

func (b *ConnectorBuilder[C, P]) Build(opts []func(params *P)) (conn *C, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	var paramsTemplate P

	params, err := paramsbuilder.Apply(paramsTemplate, opts)
	if err != nil {
		return nil, err
	}

	clients, err := internalNewClients(b.provider, params)
	if err != nil {
		return nil, err
	}

	clients.WithErrorHandler(b.errorHandler)

	// TODO everything from this line till the end of method is concerned about
	// Dependency Injection
	// Connector which is build using composition must self wire
	// The starting values are either coming from
	// * options which produce params, which in turn are converted to some objects
	// * Setup some values could be coming from the implementor of deep connector rather than end user (aka options)

	var connectorTemplate C
	connector := &connectorTemplate

	a, ok := any(connector).(Assignable[Clients])
	if !ok {
		return nil, errors.New("there is no clients field to attach connector to")
	}

	a.CopyFrom(clients)

	if b.setup != nil {
		b.setup(connector)
	}

	return connector, nil
}
