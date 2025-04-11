package pipeliner

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector

	workspace string
}

// NewConnector is an old constructor, use NewConnectorV2.
// Deprecated.
func NewConnector(opts ...Option) (*Connector, error) {
	params, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	return NewConnectorV2(*params)
}

func NewConnectorV2(params common.Parameters) (*Connector, error) {
	conn, err := components.Initialize(providers.Pipeliner, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.workspace = params.Workspace

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle)

	return &Connector{Connector: base}, nil
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL(parts...)
}
