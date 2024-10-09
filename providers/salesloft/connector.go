package salesloft

import (
	"errors"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v2"

type Connector struct {
	deep.Clients
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
		clients *deep.Clients,
		closer *deep.EmptyCloser,
		reader *deep.Reader,
		writer *deep.Writer,
		staticMetadata *deep.StaticMetadata,
		remover *deep.Remover) *Connector {

		reader.SupportedReadObjects = &supportedObjectsByRead // TODO express as dependency (SupportedObjectByOperation)

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
	firstPage := deep.FirstPageBuilder{
		Build: func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
			url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

			if !config.Since.IsZero() {
				// Documentation states ISO8601, while server accepts different formats
				// but for consistency we are sticking to one format to be sent.
				// For the reference any API resource that includes time data type mentions iso8601 string format.
				// One example, say accounts is https://developers.salesloft.com/docs/api/accounts-index
				updatedSince := config.Since.Format(time.RFC3339Nano)
				url.WithQueryParam("updated_at[gte]", updatedSince)
			}

			return url, nil
		},
	}
	nextPage := deep.NextPageBuilder{
		Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error) {
			nextPageNum, err := jsonquery.New(node, "metadata", "paging").Integer("next_page", true)
			if err != nil {
				if errors.Is(err, jsonquery.ErrKeyNotFound) {
					// list resource doesn't support pagination, hence no next page
					return nil, nil
				}

				return nil, err
			}

			if nextPageNum == nil {
				// next page doesn't exist
				return nil, nil
			}

			// use request URL to infer the next page URL
			previousPage.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))

			return previousPage, nil
		},
	}
	readObjectLocator := deep.ReadObjectLocator{
		Locate: func(config common.ReadParams) string {
			return "data"
		},
	}
	urlResolver := deep.URLResolver{
		Resolve: func(baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}
	writeResultBuilder := deep.WriteResultBuilder{
		Build: func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
			nested, err := jsonquery.New(body).Object("data", false)
			if err != nil {
				return nil, err
			}

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
	meta := deep.StaticMetadataHolder{
		Metadata: metadata.Schemas,
	}

	return deep.Connector[Connector, parameters](constructor, providers.Salesloft, &errorHandler, opts,
		meta,
		urlResolver,
		firstPage,
		nextPage,
		readObjectLocator,
		writeResultBuilder,
	)
}
