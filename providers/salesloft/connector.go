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
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/providers"
	"github.com/spyzhov/ajson"
)

const apiVersion = "v2"

type Connector struct {
	deep.Clients
	deep.EmptyCloser
	deep.Reader
}

func constructor(
	clients *deep.Clients,
	closer *deep.EmptyCloser,
	reader *deep.Reader,
) *Connector {
	return &Connector{
		Clients:     *clients,
		EmptyCloser: *closer,
		Reader:      *reader,
	}
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](constructor, providers.Salesloft, opts,
		errorHandler,
		objectURLResolver,
		objectSupport,
		readFirstPage,
		readNextPage,
		readResponse,
	)
}

var (
	// Connector components.
	errorHandler = interpreter.ErrorHandler{ //nolint:gochecknoglobals
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}
	objectURLResolver = dpobjects.URLFormat{ //nolint:gochecknoglobals
		Produce: func(method dpobjects.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
			return urlbuilder.New(baseURL, apiVersion, objectName)
		},
	}
	objectSupport = dpobjects.SupportRegistry{ //nolint:gochecknoglobals
		Read: supportedObjectsByRead,
	}
	readFirstPage = dpread.FirstPageBuilder{ //nolint:gochecknoglobals
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
	readNextPage = dpread.NextPageBuilder{ //nolint:gochecknoglobals
		Build: func(config common.ReadParams, url *urlbuilder.URL, node *ajson.Node) (string, error) {
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
			url.WithQueryParam("page", strconv.FormatInt(*nextPageNum, 10))

			return url.String(), nil
		},
	}
	readResponse = dpread.ResponseLocator{ //nolint:gochecknoglobals
		Locate: func(config common.ReadParams, node *ajson.Node) string {
			return "data"
		},
	}
)

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL(), apiVersion, arg)
}
