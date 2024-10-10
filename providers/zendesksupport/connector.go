package zendesksupport

import (
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v2"

type Connector struct {
	dprequests.Clients
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

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *dprequests.Clients,
		closer *deep.EmptyCloser,
		reader *deep.Reader,
		writer *deep.Writer,
		remover *deep.Remover,
		staticMetadata *deep.StaticMetadata,
	) *Connector {
		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			Writer:         *writer,
			StaticMetadata: *staticMetadata,
			Remover:        *remover,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	objectURLResolver := dpobjects.SingleURLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}
	objectSupport := dpobjects.ObjectSupport{
		Read: supportedObjectsByRead,
	}
	nextPage := dpread.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			return jsonquery.New(node, "links").StrWithDefault("next", "")
		},
	}
	readObjectLocator := dpread.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return ObjectNameToResponseField.Get(config.ObjectName)
		},
	}
	writeResultBuilder := deep.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			nested, err := jsonquery.New(body).Object(config.ObjectName, true)
			if err != nil {
				return nil, err
			}

			if nested == nil {
				// Field should be in singular form. Either one will be matched.
				// This one is NOT optional.
				nested, err = jsonquery.New(body).Object(
					naming.NewSingularString(config.ObjectName).String(),
					false,
				)
				if err != nil {
					return nil, err
				}
			}
			// nested node now must be not null, carry on

			rawID, err := jsonquery.New(nested).Integer("id", true)
			if err != nil {
				return nil, err
			}

			recordID := ""
			if rawID != nil {
				// optional
				recordID = strconv.FormatInt(*rawID, 10)
			}

			data, err := jsonquery.Convertor.ObjectToMap(nested)
			if err != nil {
				return nil, err
			}

			return &common.WriteResult{
				Success:  true,
				RecordId: recordID,
				Errors:   nil,
				Data:     data,
			}, nil
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.ZendeskSupport, opts,
		meta,
		errorHandler,
		objectURLResolver,
		objectSupport,
		nextPage,
		readObjectLocator,
		writeResultBuilder,
	)
}
