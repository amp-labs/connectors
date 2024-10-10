package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
)

type Remover struct {
	clients       Clients
	urlResolver   URLResolver
	objectManager ObjectManager
}

func NewRemover(clients *Clients, resolver URLResolver, objectManager ObjectManager) *Remover {
	return &Remover{
		clients:       *clients,
		urlResolver:   resolver,
		objectManager: objectManager,
	}
}

func (r *Remover) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !r.objectManager.IsDeleteSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := r.urlResolver.FindURL(DeleteMethod, r.clients.BaseURL(), config.ObjectName)
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
