package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
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
	resolver *URLResolver,
	firstPageBuilder *FirstPageBuilder,
	nextPageBuilder *NextPageBuilder,
	objectLocator *ReadObjectLocator,
	objectManager ObjectManager) *Reader {
	return &Reader{
		urlResolver:       *resolver,
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

	return common.ParseResult(
		rsp,
		common.GetRecordsUnderJSONPath(r.readObjectLocator.Locate(config)),
		func(node *ajson.Node) (string, error) {
			nextURL, err := r.nextPageBuilder.Build(config, url, node)
			if err != nil {
				return "", err
			}
			if nextURL == nil {
				return "", nil
			}

			return nextURL.String(), nil
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
	url, err := r.urlResolver.Resolve(ReadMethod, r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	return r.firstPageBuilder.Build(config, url)
}

type FirstPageBuilder struct {
	Build func(config common.ReadParams, url *urlbuilder.URL) (*urlbuilder.URL, error)
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

func (b NextPageBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "nextPageBuilder",
		Constructor: handy.Returner(b),
	}
}

type ReadObjectLocator struct {
	Locate func(config common.ReadParams) string
}

func (l ReadObjectLocator) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "readObjectLocator",
		Constructor: handy.Returner(l),
	}
}
