package gong

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gong/metadata"
	"github.com/spyzhov/ajson"
)

const ApiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
	deep.Writer
	deep.StaticMetadata
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		reader *deep.Reader,
		writer *deep.Writer,
		staticMetadata *deep.StaticMetadata,
	) *Connector {
		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			Writer:         *writer,
			StaticMetadata: *staticMetadata,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	objectSupport := dpobjects.ObjectSupport{
		Read:  supportedObjectsByRead,
		Write: supportedObjectsByWrite,
	}
	objectURLResolver := dpobjects.SingleURLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, ApiVersion, objectName)
		},
	}
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			if !config.Since.IsZero() {
				// This time format is RFC3339 but in UTC only.
				// See calls or users object for query parameter requirements.
				// https://gong.app.gong.io/settings/api/documentation#get-/v2/calls
				url.WithQueryParam("fromDateTime", handy.Time.FormatRFC3339inUTC(config.Since))
			}

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			nextPageCursor, err := jsonquery.New(node, "records").StrWithDefault("cursor", "")
			if err != nil {
				return "", err
			}

			if len(nextPageCursor) != 0 {
				previousPage.WithQueryParam("cursor", nextPageCursor)

				return previousPage.String(), nil
			}

			return "", nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return config.ObjectName
		},
	}
	writeResultBuilder := deep.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			recordID, err := jsonquery.New(body).Str("callId", false)
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

	return deep.Connector[Connector, parameters](constructor, providers.Gong, opts,
		meta,
		errorHandler,
		objectSupport,
		objectURLResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		deep.PostWriteRequestBuilder{},
		writeResultBuilder,
	)
}
