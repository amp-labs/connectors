package dpread

import (
	"context"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// RequestGet performs object reading via GET operation.
// You may specify callback to enhance request with Object specific headers.
// Otherwise, consider dprequests.HeaderSupplements for constant headers.
type RequestGet struct {
	MakeHeaders func(objectName string, clients dprequests.Clients) []common.Header
}

func (r RequestGet) MakeReadRequest(
	objectName string, clients dprequests.Clients,
) (common.ReadMethod, []common.Header) {

	var headers []common.Header
	if r.MakeHeaders != nil {
		headers = r.MakeHeaders(objectName, clients)
	}

	return func(
		ctx context.Context, url *urlbuilder.URL, body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		// Wrapper around GET without request body.

		return clients.JSON.Get(ctx, url.String(), headers...)
	}, headers
}

func (r RequestGet) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ReadRequestBuilder,
		Constructor: handy.PtrReturner(r),
		Interface:   new(Requester),
	}
}
