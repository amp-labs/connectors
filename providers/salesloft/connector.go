package salesloft

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
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
	return components.Initialize(providers.Salesloft, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle)

	return &Connector{Connector: base}, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path, _ := metadata.Schemas.LookupRawURLPath(c.Module(), objectName)
	if len(path) == 0 {
		// Fallback, try objectName as a URL.
		path = objectName
	}

	return c.ModuleClient.URL(path)
}
