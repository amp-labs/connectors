package deep

import (
	"context"
	"github.com/amp-labs/connectors/internal/deep/requirements"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpremove"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
)

type Remover struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	objectManager     dpobjects.ObjectManager
	urlResolver       dpobjects.ObjectURLResolver
	requestBuilder    dpremove.RemoveRequestBuilder
}

func newRemover(clients *dprequests.Clients,
	resolver dpobjects.ObjectURLResolver,
	objectManager dpobjects.ObjectManager,
	requestBuilder dpremove.RemoveRequestBuilder,
	headerSupplements *dprequests.HeaderSupplements,
) *Remover {
	return &Remover{
		clients:           *clients,
		urlResolver:       resolver,
		objectManager:     objectManager,
		requestBuilder:    requestBuilder,
		headerSupplements: *headerSupplements,
	}
}

func (r Remover) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !r.objectManager.IsDeleteSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := r.urlResolver.FindURL(dpobjects.DeleteMethod, r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	deleteMethod, headers := r.requestBuilder.MakeDeleteRequest(config.ObjectName, config.RecordId, r.clients)
	headers = append(headers, r.headerSupplements.DeleteHeaders()...)

	_, err = deleteMethod(ctx, url, nil, headers...)
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}

func (r Remover) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:         requirements.Remover,
		Constructor: newRemover,
	}
}
