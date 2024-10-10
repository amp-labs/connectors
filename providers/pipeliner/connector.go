package pipeliner

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipeliner/metadata"
	"github.com/spyzhov/ajson"
	"strconv"
)

type Connector struct {
	Data deep.ConnectorData[parameters, *deep.EmptyMetadataVariables]
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		data *deep.ConnectorData[parameters, *deep.EmptyMetadataVariables],
		reader *deep.Reader,
		//writer *deep.Writer,
		metadata *deep.StaticMetadata,
		//remover *deep.Remover
	) *Connector {
		return &Connector{
			Data:           *data,
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			StaticMetadata: *metadata,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			url.WithQueryParam("first", strconv.Itoa(DefaultPageSize))

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error) {
			after, err := jsonquery.New(node, "page_info").StrWithDefault("end_cursor", "")
			if err != nil {
				return nil, err
			}

			if len(after) != 0 {
				previousPage.WithQueryParam("after", after)

				return previousPage, nil
			}

			return nil, nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams) string {
			return "data"
		},
	}
	objectManager := deep.ObjectRegistry{
		Read: supportedObjectsByRead,
	}

	return deep.Connector[Connector, parameters](constructor, providers.Atlassian, opts,
		errorHandler,
		meta,
		customURLBuilder{},
		firstPage,
		nextPage,
		readObjectLocator,
		objectManager,
	)
}

func (c *Connector) getURL(parts ...string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), append([]string{
		"api/v100/rest/spaces/", c.Data.Workspace, "/entities",
	}, parts...)...)
}
