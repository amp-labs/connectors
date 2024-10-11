package dpwrite

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type WriteResultBuilder struct {
	Build func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error)
}

func (b WriteResultBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteResultBuilder,
		Constructor: handy.PtrReturner(b),
	}
}

type WriteRequestBuilder interface {
	requirements.ConnectorComponent

	MakeCreateRequest(
		objectName string, url *urlbuilder.URL, clients dprequests.Clients) (common.WriteMethod, []common.Header)
	MakeUpdateRequest(
		objectName, recordID string, url *urlbuilder.URL, clients dprequests.Clients) (common.WriteMethod, []common.Header)
}

type PostPutWriteRequestBuilder struct {
	SimplePostCreateRequest
	simplePutUpdateRequest
}

var _ WriteRequestBuilder = PostPutWriteRequestBuilder{}

func (b PostPutWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostWriteRequestBuilder struct {
	SimplePostCreateRequest
	simpleNoopUpdateRequest
}

var _ WriteRequestBuilder = PostWriteRequestBuilder{}

func (b PostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostPatchWriteRequestBuilder struct {
	SimplePostCreateRequest
	SimplePatchUpdateRequest
}

var _ WriteRequestBuilder = PostPatchWriteRequestBuilder{}

func (b PostPatchWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostPostWriteRequestBuilder struct {
	SimplePostCreateRequest
	SimplePostUpdateRequest
}

var _ WriteRequestBuilder = PostPostWriteRequestBuilder{}

func (b PostPostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(WriteRequestBuilder),
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
