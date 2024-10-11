package dpwrite

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// RequestPostPut does write, where create uses POST, update uses PUT.
type RequestPostPut struct {
	CreateViaPost
	UpdateViaPut
}

// RequestPostNoop does write, where create uses POST, without update support.
type RequestPostNoop struct {
	CreateViaPost
	UpdateNoop
}

// RequestPostPatch does write, where create uses POST, update uses PATCH.
type RequestPostPatch struct {
	CreateViaPost
	UpdateViaPatch
}

// RequestPostPost does write, where create uses POST, update uses POST.
type RequestPostPost struct {
	CreateViaPost
	UpdateViaPost
}

type CreateViaPost struct{}

func (CreateViaPost) MakeCreateRequest(
	objectName string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	return clients.JSON.Post, nil
}

type UpdateViaPut struct{}

func (UpdateViaPut) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Put, nil
}

type UpdateNoop struct{}

func (UpdateNoop) MakeUpdateRequest(
	string, string, *urlbuilder.URL, dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	return nil, nil
}

type UpdateViaPatch struct{}

func (UpdateViaPatch) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Patch, nil
}

type UpdateViaPost struct{}

func (UpdateViaPost) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients dprequests.Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Post, nil
}

func (b RequestPostPut) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

func (b RequestPostNoop) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

func (b RequestPostPatch) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}

func (b RequestPostPost) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.WriteRequestBuilder,
		Constructor: handy.PtrReturner(b),
		Interface:   new(Requester),
	}
}
