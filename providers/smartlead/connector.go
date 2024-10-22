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

// Connector
// (Read)
//
//	Pagination is not supported for this provider.
//
// (Write)
//
//	Operation allows:
//	 * create campaigns
//	 * create/update email-accounts
//	 * create client
type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader

	deep.Writer
	deep.StaticMetadata
	deep.Remover
}

func constructor(
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

type parameters struct {
	paramsbuilder.Client
	// Error is set when any With<Method> fails, used for parameters validation.
	setupError error
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](constructor, providers.Smartlead, opts,
		errorHandler,
		objectURLResolver,
		objectSupport,
		readResponse,
		writeResponse,
		dpwrite.RequestPostPost{},
		metadataSchema,
	)
}

var (
	// Connector components.
	errorHandler = interpreter.ErrorHandler{ //nolint:gochecknoglobals
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
		HTML: &interpreter.DirectFaultyResponder{Callback: interpretHTMLError},
	}
	objectURLResolver = dpobjects.URLFormat{ //nolint:gochecknoglobals
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
	objectSupport = dpobjects.SupportRegistry{ //nolint:gochecknoglobals
		Read:   supportedObjectsByRead,
		Write:  supportedObjectsByWrite,
		Delete: supportedObjectsByDelete,
	}
	readResponse = dpread.ResponseLocator{ //nolint:gochecknoglobals
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			// The response is already an array. Empty string signifies to look "here" for array.
			return ""
		},
	}
	writeResponse = dpwrite.ResponseBuilder{ //nolint:gochecknoglobals
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
	metadataSchema = dpmetadata.SchemaHolder{ //nolint:gochecknoglobals
		Metadata: metadata.Schemas,
	}
)
