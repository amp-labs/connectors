package deep

import (
	"context"
	"github.com/spyzhov/ajson"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type Reader struct {
	urlResolver       URLResolver
	firstPageBuilder  FirstPageBuilder
	nextPageBuilder   NextPageBuilder
	readObjectLocator ReadObjectLocator

	clients              Clients
	SupportedReadObjects *handy.Set[string] // TODO maybe this should be a dependency
}

func NewReader(clients *Clients,
	resolver *URLResolver,
	firstPageBuilder *FirstPageBuilder,
	nextPageBuilder *NextPageBuilder,
	objectLocator *ReadObjectLocator) *Reader {
	return &Reader{
		clients:           *clients,
		urlResolver:       *resolver,
		firstPageBuilder:  *firstPageBuilder,
		nextPageBuilder:   *nextPageBuilder,
		readObjectLocator: *objectLocator,
	}
}

func (r *Reader) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if r.SupportedReadObjects != nil && !r.SupportedReadObjects.Has(config.ObjectName) {
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

	return common.ParseResult(
		rsp,
		common.GetRecordsUnderJSONPath(r.readObjectLocator.Locate(config)),
		func(node *ajson.Node) (string, error) {
			return r.nextPageBuilder.Build(config, url, node)
		},
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
	url, err := r.urlResolver.Resolve(r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	return r.firstPageBuilder.Build(config, url), nil
}

type FirstPageBuilder struct {
	Build func(config common.ReadParams, url *urlbuilder.URL) *urlbuilder.URL
}

type NextPageBuilder struct {
	Build func(config common.ReadParams, previousPage *urlbuilder.URL, node *ajson.Node) (string, error)
}

type ReadObjectLocator struct {
	Locate func(config common.ReadParams) string
}
