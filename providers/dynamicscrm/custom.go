package dynamicscrm

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpremove"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var (
	_ dpwrite.Requester  = customWriterRequestBuilder{}
	_ dpremove.Requester = customRemoveRequestBuilder{}
)

type customWriterRequestBuilder struct {
	dpwrite.CreateViaPost
}

func (customWriterRequestBuilder) MakeUpdateRequest(
	objectName, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	// Microsoft doesn't add IDs as a separate URI part.
	// It is in format: .../Orders(123)
	url.RawAddToPath(fmt.Sprintf("(%v)", recordID))

	return clients.JSON.Patch, nil
}

func (b customWriterRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(dpwrite.Requester),
	}
}

type customRemoveRequestBuilder struct{}

func (b customRemoveRequestBuilder) MakeDeleteRequest(
	objectName, recordID string, clients dprequests.Clients,
) (common.DeleteMethod, []common.Header) {
	// Wrapper around DELETE without request body.
	return func(ctx context.Context, url *urlbuilder.URL,
		body any, headers ...common.Header,
	) (*common.JSONHTTPResponse, error) {
		// Just like PATCH, the removal needs ID in brackets "(...)"
		url.RawAddToPath(fmt.Sprintf("(%v)", recordID))

		return clients.JSON.Delete(ctx, url.String(), headers...)
	}, nil
}

func (b customRemoveRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.RemoveRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(dpremove.Requester),
	}
}