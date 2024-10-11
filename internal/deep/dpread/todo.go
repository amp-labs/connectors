package dpread

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type PaginationStartBuilder interface {
	requirements.ConnectorComponent
	FirstPage(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
}

var _ PaginationStartBuilder = FirstPageBuilder{}

type FirstPageBuilder struct {
	Build func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
}

func (b FirstPageBuilder) FirstPage(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
	if b.Build == nil {
		// TODO error
		return nil, errors.New("build method cannot be empty")
	}

	return b.Build(config, url)
}

func (b FirstPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.PaginationStartBuilder,
		Constructor: handy.Returner(b),
		Interface:   new(PaginationStartBuilder),
	}
}

var _ PaginationStartBuilder = DefaultPageBuilder{}

type DefaultPageBuilder struct{}

func (b DefaultPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.PaginationStartBuilder,
		Constructor: handy.Returner(b),
		Interface:   new(PaginationStartBuilder),
	}
}

func (b DefaultPageBuilder) FirstPage(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
	return url, nil
}

type NextPageBuilder struct {
	Build func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error)
}

func (b NextPageBuilder) GetNextPageFunc(config common.ReadParams, url *urlbuilder.URL) (common.NextPageFunc, error) {
	if b.Build == nil {
		// TODO error
		return nil, errors.New("build method cannot be empty")
	}

	return func(node *ajson.Node) (string, error) {
		return b.Build(config, url, node)
	}, nil
}

func (b NextPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.NextPageBuilder,
		Constructor: handy.Returner(b),
	}
}

type ReadObjectLocator struct {
	// Locate should return the fieldName where desired list of Objects is located.
	Locate func(config common.ReadParams, node *ajson.Node) string
	// FlattenRecords is optional and will be used after list was located and extra processing is needed.
	// The desired fields could be nested
	FlattenRecords func(arr []*ajson.Node) ([]map[string]any, error)
}

func (l ReadObjectLocator) GetRecordsFunc(config common.ReadParams) (common.RecordsFunc, error) {
	if l.Locate == nil {
		// TODO error
		return nil, errors.New("locate method cannot be empty")
	}

	return func(node *ajson.Node) ([]map[string]any, error) {
		fieldName := l.Locate(config, node)

		arr, err := jsonquery.New(node).Array(fieldName, false)
		if err != nil {
			return nil, err
		}

		if l.FlattenRecords != nil {
			return l.FlattenRecords(arr)
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}, nil
}

func (l ReadObjectLocator) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ReadObjectLocator,
		Constructor: handy.Returner(l),
	}
}

type ReadRequestBuilder interface {
	requirements.ConnectorComponent

	MakeReadRequest(objectName string, clients dprequests.Clients) (common.ReadMethod, []common.Header)
}

var _ ReadRequestBuilder = GetRequestBuilder{}

type GetRequestBuilder struct {
	simpleGetReadRequest
}

func (b GetRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ReadRequestBuilder,
		Constructor: handy.Returner(b),
		Interface:   new(ReadRequestBuilder),
	}
}

var _ ReadRequestBuilder = GetWithHeadersRequestBuilder{}

type GetWithHeadersRequestBuilder struct {
	delegate simpleGetReadRequest
	Headers  []common.Header
}

func (b GetWithHeadersRequestBuilder) MakeReadRequest(
	objectName string, clients dprequests.Clients,
) (common.ReadMethod, []common.Header) {
	method, _ := b.delegate.MakeReadRequest(objectName, clients)

	return method, b.Headers
}

func (b GetWithHeadersRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ReadRequestBuilder,
		Constructor: handy.Returner(b),
		Interface:   new(ReadRequestBuilder),
	}
}

type simpleGetReadRequest struct{}

func (simpleGetReadRequest) MakeReadRequest(
	objectName string, clients dprequests.Clients,
) (common.ReadMethod, []common.Header) {
	// Wrapper around GET without request body.
	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		return clients.JSON.Get(ctx, url.String(), headers...)
	}, nil
}
