package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		reader *deep.Reader,
		staticMetadata *deep.StaticMetadata,
	) *Connector {
		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			StaticMetadata: *staticMetadata,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	objectURLResolver := deep.SingleURLFormat{
		Produce: func(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}
	objectSupport := deep.ObjectSupport{
		Read: supportedObjectsByRead,
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			return jsonquery.New(node, "links").StrWithDefault("next", "")
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return ObjectNameToResponseField.Get(config.ObjectName)
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.ZendeskSupport, opts,
		meta,
		errorHandler,
		objectURLResolver,
		objectSupport,
		nextPage,
		readObjectLocator,
	)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
