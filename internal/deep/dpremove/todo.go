package dpremove

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type RemoveRequestBuilder interface {
	requirements.ConnectorComponent

	MakeDeleteRequest(objectName, recordID string, clients dprequests.Clients) (common.DeleteMethod, []common.Header)
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

type simpleRemoveDeleteRequest struct{}

func (simpleRemoveDeleteRequest) MakeDeleteRequest(objectName, recordID string, clients dprequests.Clients) (common.DeleteMethod, []common.Header) {
	// Wrapper around DELETE without request body.
	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		url.AddPath(recordID)

		return clients.JSON.Delete(ctx, url.String(), headers...)
	}, nil
}
