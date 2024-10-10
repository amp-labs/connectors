package smartlead

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
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
	// Write method allows to
	// * create campaigns
	// * create/update email-accounts
	// * create client
	deep.Writer
	deep.StaticMetadata
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
		writer *deep.Writer,
		staticMetadata *deep.StaticMetadata,
		remover *deep.Remover,
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
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	objectURLResolver := deep.SingleURLFormat{
		Produce: func(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			var path string
			switch method {
			case deep.ReadMethod:
				path = objectName
			case deep.CreateMethod:
				path = createObjects[objectName]
			case deep.UpdateMethod:
				path = updateObjects[objectName]
			case deep.DeleteMethod:
				path = objectName
			}

			return urlbuilder.New(baseURL, apiVersion, path)
		},
	}
	objectManager := deep.ObjectRegistry{
		Read:   supportedObjectsByRead,
		Write:  supportedObjectsByWrite,
		Delete: supportedObjectsByDelete,
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			// Pagination is not supported for this provider.
			return "", nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			// The response is already an array. Empty string signifies to look "here" for array.
			return ""
		},
	}
	writeResultBuilder := deep.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			recordIdNodePath := writeResponseRecordIdPaths[config.ObjectName]

			// ID is integer that is always stored under different field name.
			rawID, err := jsonquery.New(body).Integer(recordIdNodePath, true)
			if err != nil {
				return nil, err
			}

			recordID := ""
			if rawID != nil {
				// optional
				recordID = strconv.FormatInt(*rawID, 10)
			}

			return &common.WriteResult{
				Success:  true,
				RecordId: recordID,
				Errors:   nil,
				Data:     nil,
			}, nil
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Smartlead, opts,
		meta,
		errorHandler,
		objectURLResolver,
		objectManager,
		firstPage,
		nextPage,
		readObjectLocator,
		writeResultBuilder,
		deep.PostPostWriteRequestBuilder{},
	)
}
