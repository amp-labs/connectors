package instantly

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/instantly/metadata"
	"github.com/spyzhov/ajson"
	"strconv"
)

const apiVersion = "v1"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.StaticMetadata
	// Delete removes object. As of now only removal of Tags are allowed because
	// deletion of other object types require a request payload to be added
	// c.Client.Delete does not yet support this.
	deep.Remover
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
		staticMetadata *deep.StaticMetadata,
		remover *deep.Remover) *Connector {
		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			StaticMetadata: *staticMetadata,
			Remover:        *remover,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	urlResolver := deep.URLResolver{
		Resolve: func(baseURL, objectName string) (*urlbuilder.URL, error) {
			path := objectResolver[objectName].URLPath

			return urlbuilder.New(baseURL, apiVersion, path)
		},
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			url.WithQueryParam("skip", "0")
			url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error) {
			previousStart := 0

			skipQP, ok := previousPage.GetFirstQueryParam("skip")
			if ok {
				// Try to use previous "skip" parameter to determine the next skip.
				skipNum, err := strconv.Atoi(skipQP)
				if err == nil {
					previousStart = skipNum
				}
			}

			nextStart := previousStart + DefaultPageSize
			previousPage.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
			previousPage.WithQueryParam("skip", strconv.Itoa(nextStart))

			return previousPage, nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams) string {
			return objectResolver[config.ObjectName].NodePath
		},
	}
	objectManager := deep.ObjectRegistry{
		Read:   supportedObjectsByRead,
		Delete: supportedObjectsByDelete,
	}

	return deep.Connector[Connector, parameters](constructor, providers.Instantly, opts,
		meta,
		errorHandler,
		urlResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		objectManager,
	)
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), append([]string{
		apiVersion,
	}, parts...)...)
}
