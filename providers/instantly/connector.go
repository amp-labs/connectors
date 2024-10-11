package instantly

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
	"github.com/amp-labs/connectors/providers/instantly/metadata"
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
	}
	meta := dpmetadata.SchemaHolder{
		Metadata: metadata.Schemas,
	}
	objectURLResolver := dpobjects.URLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			var path string
			switch method {
			case dpobjects.ReadMethod:
				path = readObjects[objectName].URLPath
			case dpobjects.CreateMethod:
				path = createObjects[objectName]
			case dpobjects.UpdateMethod:
				path = updateObjects[objectName]
			case dpobjects.DeleteMethod:
				path = deleteObjects[objectName]
			}

			return urlbuilder.New(baseURL, apiVersion, path)
		},
	}
	firstPage := dpread.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			url.WithQueryParam("skip", "0")
			url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

			return url, nil
		},
	}
	nextPage := dpread.NextPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error) {
			previousStart := 0

			skipQP, ok := url.GetFirstQueryParam("skip")
			if ok {
				// Try to use previous "skip" parameter to determine the next skip.
				skipNum, err := strconv.Atoi(skipQP)
				if err == nil {
					previousStart = skipNum
				}
			}

			nextStart := previousStart + DefaultPageSize
			url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
			url.WithQueryParam("skip", strconv.Itoa(nextStart))

			return url.String(), nil
		},
	}
	readObjectLocator := dpread.ResponseLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return readObjects[config.ObjectName].NodePath
		},
	}
	objectSupport := dpobjects.Registry{
		Read:   supportedObjectsByRead,
		Write:  supportedObjectsByWrite,
		Delete: supportedObjectsByDelete,
	}
	writeResultBuilder := dpwrite.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			recordIdNodePath := writeResponseRecordIdPaths[config.ObjectName]

			if recordIdNodePath == nil {
				// ID is not present inside response. Therefore, empty.
				return &common.WriteResult{
					Success:  true,
					RecordId: "",
					Errors:   nil,
					Data:     nil,
				}, nil
			}

			// ID is integer that is always stored under different field name.
			recordID, err := jsonquery.New(body).Str(*recordIdNodePath, false)
			if err != nil {
				return nil, err
			}

			return &common.WriteResult{
				Success:  true,
				RecordId: *recordID,
				Errors:   nil,
				Data:     nil,
			}, nil
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Instantly, opts,
		meta,
		errorHandler,
		objectURLResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		objectSupport,
		dpwrite.PostPatchWriteRequestBuilder{},
		writeResultBuilder,
	)
}
