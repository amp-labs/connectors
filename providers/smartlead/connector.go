package smartlead

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/smartlead/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v1"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.StaticMetadata
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
		reader *deep.Reader,
		staticMetadata *deep.StaticMetadata) *Connector {
		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			StaticMetadata: *staticMetadata,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	urlResolver := deep.URLResolver{
		Resolve: func(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			var path string
			switch method {
			case deep.ReadMethod:
				path = objectName
			}

			return urlbuilder.New(baseURL, apiVersion, path)
		},
	}
	objectManager := deep.ObjectRegistry{
		Read:   supportedObjectsByRead,
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error) {
			// Pagination is not supported for this provider.
			return nil, nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams) string {
			// The response is already an array. Empty string signifies to look "here" for array.
			return ""
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Smartlead, opts,
		meta,
		errorHandler,
		urlResolver,
		objectManager,
		firstPage,
		nextPage,
		readObjectLocator,
	)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
