package deep

import (
	"context"
	"github.com/amp-labs/connectors/internal/deep/requirements"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpremove"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
)

// Remover is a major connector component which provides Delete functionality.
// Embed this into connector struct.
// Provide dpobjects.URLResolver into deep.Connector.
type Remover struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	objectManager     dpobjects.Support
	urlResolver       dpobjects.URLResolver
	requestBuilder    dpremove.Requester
}

func newRemover(clients *dprequests.Clients,
	resolver dpobjects.URLResolver,
	objectManager dpobjects.Support,
	requestBuilder dpremove.Requester,
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
		ID:          requirements.Remover,
		Constructor: newRemover,
	}
}
