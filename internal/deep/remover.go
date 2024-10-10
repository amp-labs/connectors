package deep

import (
	"context"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type Remover struct {
	clients        Clients
	urlResolver    URLResolver
	objectManager  ObjectManager
	requestBuilder RemoveRequestBuilder
}

func NewRemover(clients *Clients,
	resolver URLResolver,
	objectManager ObjectManager,
	requestBuilder RemoveRequestBuilder) *Remover {
	return &Remover{
		clients:        *clients,
		urlResolver:    resolver,
		objectManager:  objectManager,
		requestBuilder: requestBuilder,
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

	deleteMethod, headers := r.requestBuilder.MakeDeleteRequest(config.ObjectName, config.RecordId, r.clients)

	_, err = deleteMethod(ctx, url, nil, headers...)
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}

type RemoveRequestBuilder interface {
	requirements.Requirement

	MakeDeleteRequest(objectName, recordID string, clients Clients) (common.DeleteMethod, []common.Header)
}

var _ RemoveRequestBuilder = DeleteRequestBuilder{}

type DeleteRequestBuilder struct {
	simpleRemoveDeleteRequest
}

func (b DeleteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "removeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(RemoveRequestBuilder),
	}
}

type simpleRemoveDeleteRequest struct {}

func (simpleRemoveDeleteRequest) MakeDeleteRequest(objectName, recordID string, clients Clients) (common.DeleteMethod, []common.Header) {
	// Wrapper around DELETE without request body.

	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		url.AddPath(recordID)

		return clients.JSON.Delete(ctx, url.String(), headers...)
	}, nil
}
