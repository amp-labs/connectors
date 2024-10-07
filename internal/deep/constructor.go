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
}

func Connector[C any, P paramsbuilder.ParamAssurance](
	provider providers.Provider,
	errorHandler interpreter.ErrorHandler) *ConnectorBuilder[C, P] {
	return &ConnectorBuilder[C, P]{
		provider:     provider,
		errorHandler: errorHandler,
	}
}

func (b ConnectorBuilder[C, P]) Build(opts []func(params *P)) (conn *C, outErr error) {
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

	var connectorTemplate C
	connector := &connectorTemplate

	a, ok := any(connector).(Assignable[Clients])
	if !ok {
		return nil, errors.New("there is no clients field to attach connector to")
	}

	a.CopyFrom(clients)

	return connector, nil
}
