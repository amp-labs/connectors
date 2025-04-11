package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
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
	return components.Initialize(providers.Klaviyo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		Custom: map[string]interpreter.FaultyResponseHandler{
			"application/vnd.api+json": interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
		},
	}.Handle)

	return &Connector{Connector: base}, nil
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupRawURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return c.ModuleClient.URL("api/", path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return c.ModuleClient.URL(path)
}

func (c *Connector) getDeleteURL(objectName string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("api/", objectName)
}

func (c *Connector) revisionHeader() common.Header {
	return common.Header{
		Key:   "revision",
		Value: string(c.Module()),
	}
}
