package intercom

import (
	"errors"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/intercom/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "2.11"

var apiVersionHeader = common.Header{ // nolint:gochecknoglobals
	Key:   "Intercom-Version",
	Value: apiVersion,
}

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
}

func NewConnector(opts ...Option) (*Connector, error) {
	constructor := func(
		clients *dprequests.Clients,
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
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}
	headerSupplements := dprequests.HeaderSupplements{
		All: []common.Header{
			apiVersionHeader,
		},
	}
	objectSupport := dpobjects.ObjectSupport{
		Read: supportedObjectsByRead,
	}
	objectURLResolver := dpobjects.SingleURLFormat{
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			url, err := urlbuilder.New(baseURL, objectName)
			if err != nil {
				return nil, err
			}

			// Intercom pagination cursor sometimes ends with `=`.
			url.AddEncodingExceptions(map[string]string{ //nolint:gochecknoglobals
				"%3D": "=",
			})

			return url, nil
		},
	}
	firstPage := dpread.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

			return url, nil
		},
	}
	nextPage := dpread.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
			next, err := jsonquery.New(node, "pages").StrWithDefault("next", "")
			if err == nil {
				return next, nil
			}

			if !errors.Is(err, jsonquery.ErrNotString) {
				// response from server doesn't meet any format that we expect
				return "", err
			}

			// Probably, we are dealing with an object under `pages.next`
			startingAfter, err := jsonquery.New(node, "pages", "next").Str("starting_after", true)
			if err != nil {
				return "", err
			}

			if startingAfter == nil {
				// next page doesn't exist
				return "", nil
			}

			previousPage.WithQueryParam("starting_after", *startingAfter)

			return previousPage.String(), nil
		},
	}
	readObjectLocator := dpread.ReadObjectLocator{
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return extractListFieldName(node)
		},
	}
	writeResultBuilder := dpwrite.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			recordID, err := jsonquery.New(body).StrWithDefault("id", "")
			if err != nil {
				return nil, err
			}

			data, err := jsonquery.Convertor.ObjectToMap(body)
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

	return deep.Connector[Connector, parameters](constructor, providers.Intercom, opts,
		meta,
		errorHandler,
		headerSupplements,
		objectSupport,
		objectURLResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		writeResultBuilder,
	)
}

func (c *Connector) getURL(objectName string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.Clients.BaseURL(), objectName)
	if err != nil {
		return nil, err
	}

	// Intercom pagination cursor sometimes ends with `=`.
	url.AddEncodingExceptions(map[string]string{ //nolint:gochecknoglobals
		"%3D": "=",
	})

	return url, nil
}
