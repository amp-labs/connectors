package pipeliner

import (
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipeliner/metadata"
	"github.com/spyzhov/ajson"
)

type Connector struct {
	Data dpvars.ConnectorData[parameters, *dpvars.EmptyMetadataVariables]
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.Writer
	deep.StaticMetadata
	deep.Remover
}

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		data *dpvars.ConnectorData[parameters, *dpvars.EmptyMetadataVariables],
		reader *deep.Reader,
		writer *deep.Writer,
		metadata *deep.StaticMetadata,
		remover *deep.Remover,
	) *Connector {
		return &Connector{
			Data:           *data,
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			Writer:         *writer,
			StaticMetadata: *metadata,
			Remover:        *remover,
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
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			after, err := jsonquery.New(node, "page_info").StrWithDefault("end_cursor", "")
			if err != nil {
				return "", err
			}

			if len(after) != 0 {
				previousPage.WithQueryParam("after", after)

				return previousPage.String(), nil
			}

			return "", nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return "data"
		},
	}
	objectSupport := dpobjects.ObjectSupport{
		Read: supportedObjectsByRead,
	}
	writeResultBuilder := deep.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			success, err := jsonquery.New(body).Bool("success", false)
			if err != nil {
				return nil, err
			}

			nested, err := jsonquery.New(body).Object("data", false)
			if err != nil {
				return nil, err
			}

			recordID, err := jsonquery.New(nested).StrWithDefault("id", "")
			if err != nil {
				return nil, err
			}

			data, err := jsonquery.Convertor.ObjectToMap(nested)
			if err != nil {
				return nil, err
			}

			return &common.WriteResult{
				Success:  *success,
				RecordId: recordID,
				Errors:   nil,
				Data:     data,
			}, nil
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Atlassian, opts,
		errorHandler,
		meta,
		customURLBuilder{},
		firstPage,
		nextPage,
		readObjectLocator,
		objectSupport,
		deep.PostPatchWriteRequestBuilder{},
		writeResultBuilder,
	)
}
