package salesloft

import (
	"errors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/spyzhov/ajson"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
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
		staticMetadata *deep.StaticMetadata,
		remover *deep.Remover) *Connector {

		reader.SupportedReadObjects = &supportedObjectsByRead // TODO dependency in some way

		return &Connector{
			Clients:        *clients,
			EmptyCloser:    *closer,
			Reader:         *reader,
			StaticMetadata: *staticMetadata,
			Remover:        *remover,
		}
	}
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	urlResolver := &deep.URLResolver{
		Resolve: func(baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}

	return deep.Connector[Connector, parameters](constructor, providers.Salesloft, &errorHandler, opts,
		deep.Dependency{
			Constructor: func() *scrapper.ObjectMetadataResult {
				return metadata.Schemas
			},
		},
		deep.Dependency{
			Constructor: func() *deep.URLResolver {
				return urlResolver
			},
		},
		deep.Dependency{
			Constructor: func() *deep.FirstPageBuilder {
				return &deep.FirstPageBuilder{
					Build: func(config common.ReadParams, url *urlbuilder.URL) *urlbuilder.URL {
						url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

						if !config.Since.IsZero() {
							// Documentation states ISO8601, while server accepts different formats
							// but for consistency we are sticking to one format to be sent.
							// For the reference any API resource that includes time data type mentions iso8601 string format.
							// One example, say accounts is https://developers.salesloft.com/docs/api/accounts-index
							updatedSince := config.Since.Format(time.RFC3339Nano)
							url.WithQueryParam("updated_at[gte]", updatedSince)
						}

						return url
					},
				}
			},
		},
		deep.Dependency{
			Constructor: func() *deep.NextPageBuilder {
				return &deep.NextPageBuilder{
					Build: func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error) {
						nextPageNum, err := jsonquery.New(node, "metadata", "paging").Integer("next_page", true)
						if err != nil {
							if errors.Is(err, jsonquery.ErrKeyNotFound) {
								// list resource doesn't support pagination, hence no next page
								return "", nil
							}

							return "", err
						}

						if nextPageNum == nil {
							// next page doesn't exist
							return "", nil
						}

						// use request URL to infer the next page URL
						previousPage.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))

						return previousPage.String(), nil
					},
				}
			},
		},
		deep.Dependency{
			Constructor: func() *deep.ReadObjectLocator {
				return &deep.ReadObjectLocator{
					Locate: func(config common.ReadParams) string {
						return "data"
					},
				}
			},
		},
	)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
