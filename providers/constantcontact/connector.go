package constantcontact

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/constantcontact/metadata"
)

const ApiVersion = "v3"

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
	return components.Initialize(providers.ConstantContact, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle)

	return &Connector{Connector: base}, nil
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(c.Module(), objectName)
	if err != nil {
		var ok bool
		if path, ok = objectNameToWritePath[objectName]; !ok {
			// check out if object name is part of exceptions
			return nil, common.ErrOperationNotSupportedForObject
		}
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, ApiVersion, path)
}
