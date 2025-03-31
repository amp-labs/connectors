package iterable

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/iterable/metadata"
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
	return components.Initialize(providers.Iterable, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	conn := &Connector{Connector: base}

	conn.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: conn.interpretHTMLError},
	}.Handle)

	return conn, nil
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := supportedObjectsByWrite[objectName]

	return urlbuilder.New(c.ProviderInfo().BaseURL, path)
}

func (c *Connector) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	path := supportedObjectsByDelete[objectName]

	return urlbuilder.New(c.ProviderInfo().BaseURL, path, recordID)
}
