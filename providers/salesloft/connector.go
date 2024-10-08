package salesloft

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v2"

type Connector struct {
	*deep.Clients
	*deep.EmptyCloser
	*deep.StaticMetadata
	*deep.Remover
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		staticMetadata *deep.StaticMetadata,
		remover *deep.Remover) *Connector {
		return &Connector{
			Clients:        clients,
			EmptyCloser:    closer,
			StaticMetadata: staticMetadata,
			Remover:        remover,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	urlResolver := &deep.DirectURLResolver{
		Resolve: func(baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Salesloft, &errorHandler, opts,
		deep.Dependency{
			Constructor: func() *scrapper.ObjectMetadataResult {
				return metadata.Schemas
			},
		},
		deep.Dependency{
			Constructor: func() deep.URLResolver {
				return urlResolver
			},
		},
	)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
