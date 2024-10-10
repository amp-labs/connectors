package deep

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type Reader struct {
	urlResolver       URLResolver
	firstPageBuilder  FirstPageBuilder
	nextPageBuilder   NextPageBuilder
	readObjectLocator ReadObjectLocator
	objectManager     ObjectManager
	requestBuilder    ReadRequestBuilder

	clients Clients
}

func NewReader(clients *Clients,
	resolver URLResolver,
	firstPageBuilder *FirstPageBuilder,
	nextPageBuilder *NextPageBuilder,
	objectLocator *ReadObjectLocator,
	objectManager ObjectManager,
	requestBuilder ReadRequestBuilder,
) *Reader {
	return &Reader{
		urlResolver:       resolver,
		firstPageBuilder:  *firstPageBuilder,
		nextPageBuilder:   *nextPageBuilder,
		readObjectLocator: *objectLocator,
		objectManager:     objectManager,
		clients:           *clients,
		requestBuilder:    requestBuilder,
	}
}

func (r *Reader) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !r.objectManager.IsReadSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := r.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	read, headers := r.requestBuilder.MakeReadRequest(config.ObjectName, r.clients)

	rsp, err := read(ctx, url, nil, headers...)
	if err != nil {
		return nil, err
	}

	recordsFunc, err := r.readObjectLocator.getRecordsFunc(config)
	if err != nil {
		return nil, err
	}

	nextPageFunc, err := r.nextPageBuilder.getNextPageFunc(config, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		recordsFunc,
		nextPageFunc,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (r *Reader) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := r.urlResolver.FindURL(ReadMethod, r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	return r.firstPageBuilder.produceURL(config, url)
}

type FirstPageBuilder struct {
	Build func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
}

func (b FirstPageBuilder) produceURL(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error) {
	if b.Build == nil {
		// TODO error
		return nil, errors.New("build method cannot be empty")
	}

	return b.Build(config, url)
}

func (b FirstPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "firstPageBuilder",
		Constructor: handy.Returner(b),
	}
}

type NextPageBuilder struct {
	Build func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error)
}

func (b NextPageBuilder) getNextPageFunc(config common.ReadParams, url *urlbuilder.URL) (common.NextPageFunc, error) {
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
		ID:          "nextPageBuilder",
		Constructor: handy.Returner(b),
	}
}

type ReadObjectLocator struct {
	// Locate should return the fieldName where desired list of Objects is located.
	Locate func(config common.ReadParams) string
	// FlattenRecords is optional and will be used after list was located and extra processing is needed.
	// The desired fields could be nested
	FlattenRecords func(arr []*ajson.Node) ([]map[string]any, error)
}

func (l ReadObjectLocator) getRecordsFunc(config common.ReadParams) (common.RecordsFunc, error) {
	if l.Locate == nil {
		// TODO error
		return nil, errors.New("locate method cannot be empty")
	}

	fieldName := l.Locate(config)

	if l.FlattenRecords == nil {
		return common.GetRecordsUnderJSONPath(fieldName), nil
	}

	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := jsonquery.New(node).Array(fieldName, false)
		if err != nil {
			return nil, err
		}

		return l.FlattenRecords(arr)
	}, nil
}

func (l ReadObjectLocator) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "readObjectLocator",
		Constructor: handy.Returner(l),
	}
}

type ReadRequestBuilder interface {
	requirements.Requirement

	MakeReadRequest(objectName string, clients Clients) (common.ReadMethod, []common.Header)
}

var _ ReadRequestBuilder = GetRequestBuilder{}

type GetRequestBuilder struct {
	simpleGetReadRequest
}

func (b GetRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "readRequestBuilder",
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
	objectName string, clients Clients,
) (common.ReadMethod, []common.Header) {
	method, _ := b.delegate.MakeReadRequest(objectName, clients)

	return method, b.Headers
}

func (b GetWithHeadersRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "readRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(ReadRequestBuilder),
	}
}

type simpleGetReadRequest struct{}

func (simpleGetReadRequest) MakeReadRequest(
	objectName string, clients Clients,
) (common.ReadMethod, []common.Header) {
	// Wrapper around GET without request body.
	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		return clients.JSON.Get(ctx, url.String(), headers...)
	}, nil
}
