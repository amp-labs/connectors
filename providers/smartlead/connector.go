package smartlead

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/dpmetadata"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/smartlead/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v1"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	// Pagination is not supported for this provider.
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
	meta := dpmetadata.SchemaHolder{
		Metadata: metadata.Schemas,
	}
	objectURLResolver := dpobjects.URLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			var path string
			switch method {
			case dpobjects.ReadMethod:
				path = objectName
			case dpobjects.CreateMethod:
				path = createObjects[objectName]
			case dpobjects.UpdateMethod:
				path = updateObjects[objectName]
			case dpobjects.DeleteMethod:
				path = objectName
			}

			return urlbuilder.New(baseURL, apiVersion, path)
		},
	}
	objectSupport := dpobjects.Registry{
		Read:   supportedObjectsByRead,
		Write:  supportedObjectsByWrite,
		Delete: supportedObjectsByDelete,
	}
	readObjectLocator := dpread.ResponseLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			// The response is already an array. Empty string signifies to look "here" for array.
			return ""
		},
	}
	writeResultBuilder := dpwrite.ResponseBuilder{
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
		objectSupport,
		readObjectLocator,
		writeResultBuilder,
		dpwrite.RequestPostPost{},
	)
}
