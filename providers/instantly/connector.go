package instantly

import (
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers/instantly/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Connector struct {
	*deep.Clients
	*deep.EmptyCloser
	*deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
	// Error is set when any With<Method> fails, used for parameters validation.
	setupError error
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		staticMetadata *deep.StaticMetadata) *Connector {
		return &Connector{
			Clients:        clients,
			EmptyCloser:    closer,
			StaticMetadata: staticMetadata,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}

	return deep.Connector[Connector, parameters](constructor, providers.Instantly, &errorHandler, opts,
		deep.Dependency{
			Constructor: func() *scrapper.ObjectMetadataResult {
				return metadata.Schemas
			},
		},
	)
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), append([]string{
		apiVersion,
	}, parts...)...)
}
