package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type Writer struct {
	urlResolver   URLResolver
	resultBuilder WriteResultBuilder

	clients Clients
}

func NewWriter(clients *Clients,
	resolver *URLResolver,
	resultBuilder *WriteResultBuilder) *Writer {
	return &Writer{
		clients:       *clients,
		urlResolver:   *resolver,
		resultBuilder: *resultBuilder,
	}
}

func (w *Writer) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := w.urlResolver.Resolve(w.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = w.clients.JSON.Post
	} else {
		// only put is supported for updating 'Single' resource
		write = w.clients.JSON.Put

		url.AddPath(config.RecordId)
	}

	res, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return w.resultBuilder.Build(config, body)
}

type WriteResultBuilder struct {
	Build func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error)
}

func (b WriteResultBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeResultBuilder",
		Constructor: handy.Returner(b),
	}
}
