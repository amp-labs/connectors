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

	clients Clients
}

func NewReader(clients *Clients,
	resolver URLResolver,
	firstPageBuilder *FirstPageBuilder,
	nextPageBuilder *NextPageBuilder,
	objectLocator *ReadObjectLocator,
	objectManager ObjectManager) *Reader {
	return &Reader{
		urlResolver:       resolver,
		firstPageBuilder:  *firstPageBuilder,
		nextPageBuilder:   *nextPageBuilder,
		readObjectLocator: *objectLocator,
		objectManager:     objectManager,
		clients:           *clients,
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

	rsp, err := r.clients.JSON.Get(ctx, url.String())
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
	Build func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (*urlbuilder.URL, error)
}

func (b NextPageBuilder) getNextPageFunc(config common.ReadParams, url *urlbuilder.URL) (common.NextPageFunc, error) {
	if b.Build == nil {
		// TODO error
		return nil, errors.New("build method cannot be empty")
	}

	return func(node *ajson.Node) (string, error) {
		nextURL, err := b.Build(config, url, node)
		if err != nil {
			return "", err
		}
		if nextURL == nil {
			return "", nil
		}

		return nextURL.String(), nil
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
