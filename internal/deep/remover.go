package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
)

type Remover struct {
	clients     Clients
	urlResolver URLResolver
}

func NewRemover(clients *Clients, resolver *URLResolver) *Remover {
	return &Remover{
		clients:     *clients,
		urlResolver: *resolver,
	}
}

func (r *Remover) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := r.urlResolver.Resolve(r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	_, err = r.clients.JSON.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
