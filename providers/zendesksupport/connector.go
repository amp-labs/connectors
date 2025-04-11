package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

type Connector struct {
	// Basic connector
	*components.Connector
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
	return components.Initialize(providers.ZendeskSupport, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle)

	return &Connector{Connector: base}, nil
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupRawURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return c.ModuleClient.URL(path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	if objectsUnsupportedWrite[c.Module()].Has(objectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if path, ok := writeURLExceptions[c.Module()][objectName]; ok {
		// URL for write differs from read.
		return c.ModuleClient.URL(path)
	}

	return c.getReadURL(objectName)
}
