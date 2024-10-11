package dpwrite

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type PostPutWriteRequestBuilder struct {
	SimplePostCreateRequest
	simplePutUpdateRequest
}

func (b PostPutWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

type PostWriteRequestBuilder struct {
	SimplePostCreateRequest
	simpleNoopUpdateRequest
}

func (b PostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

type PostPatchWriteRequestBuilder struct {
	SimplePostCreateRequest
	SimplePatchUpdateRequest
}

func (b PostPatchWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

type PostPostWriteRequestBuilder struct {
	SimplePostCreateRequest
	SimplePostUpdateRequest
}

func (b PostPostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

type SimplePostCreateRequest struct{}

func (SimplePostCreateRequest) MakeCreateRequest(
	objectName string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	return clients.JSON.Post, nil
}

type simplePutUpdateRequest struct{}

func (simplePutUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Put, nil
}

type simpleNoopUpdateRequest struct{}

func (simpleNoopUpdateRequest) MakeUpdateRequest(
	string, string, *urlbuilder.URL, dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	return nil, nil
}

type SimplePatchUpdateRequest struct{}

func (SimplePatchUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Patch, nil
}

type SimplePostUpdateRequest struct{}

func (SimplePostUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Post, nil
}
