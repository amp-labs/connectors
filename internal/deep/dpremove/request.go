package dpremove

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// RequestDelete performs object removal via DELETE operation.
// You may specify callback to enhance request with Object specific headers.
// Otherwise, consider dprequests.HeaderSupplements for constant headers.
type RequestDelete struct {
	MakeHeaders func(objectName string, clients dprequests.Clients) []common.Header
}

func (d RequestDelete) MakeDeleteRequest(
	objectName, recordID string, clients dprequests.Clients,
) (common.DeleteMethod, []common.Header) {
	var headers []common.Header
	if d.MakeHeaders != nil {
		headers = d.MakeHeaders(objectName, clients)
	}

	// Wrapper around DELETE without request body.
	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		url.AddPath(recordID)

		return clients.JSON.Delete(ctx, url.String(), headers...)
	}, headers
}

func (d RequestDelete) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.RemoveRequestBuilder,
		Constructor: handy.PtrReturner(d),
		Interface:   new(Requester),
	}
}
