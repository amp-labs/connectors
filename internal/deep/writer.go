package deep

import (
	"context"
	"errors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type Writer struct {
	urlResolver    URLResolver
	resultBuilder  WriteResultBuilder
	objectManager  ObjectManager
	requestBuilder WriteRequestBuilder

	clients Clients
}

func NewWriter(clients *Clients,
	resolver *URLResolver,
	requestBuilder WriteRequestBuilder,
	resultBuilder *WriteResultBuilder,
	objectManager ObjectManager) *Writer {
	return &Writer{
		urlResolver:    *resolver,
		resultBuilder:  *resultBuilder,
		objectManager:  objectManager,
		requestBuilder: requestBuilder,
		clients:        *clients,
	}
}

func (w *Writer) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !w.objectManager.IsWriteSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	method := CreateMethod
	if len(config.RecordId) != 0 {
		method = UpdateMethod
	}

	url, err := w.urlResolver.Resolve(method, w.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	var headers []common.Header
	if len(config.RecordId) == 0 {
		write, headers = w.requestBuilder.MakeCreateRequest(config.ObjectName, url, w.clients)
	} else {
		write, headers = w.requestBuilder.MakeUpdateRequest(config.ObjectName, config.RecordId, url, w.clients)
		if write == nil {
			// TODO need a better error
			return nil, errors.New("update is not supported for this object")
		}
	}

	res, err := write(ctx, url.String(), config.RecordData, headers...)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return w.resultBuilder.Build(config, body)
}

type WriteResultBuilder struct {
	Build func(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error)
}

func (b WriteResultBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeResultBuilder",
		Constructor: handy.Returner(b),
	}
}

type WriteRequestBuilder interface {
	requirements.Requirement

	MakeCreateRequest(
		objectName string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header)
	MakeUpdateRequest(
		objectName, recordID string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header)
}

type PostPutWriteRequestBuilder struct {
	simplePostCreateRequest
	simplePutUpdateRequest
}

var _ WriteRequestBuilder = PostPutWriteRequestBuilder{}

func (b PostPutWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostWriteRequestBuilder struct{
	simplePostCreateRequest
	simpleNoopUpdateRequest
}

var _ WriteRequestBuilder = PostWriteRequestBuilder{}

func (b PostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostPatchWriteRequestBuilder struct{
	simplePostCreateRequest
	simplePatchUpdateRequest
}

var _ WriteRequestBuilder = PostPatchWriteRequestBuilder{}

func (b PostPatchWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type PostPostWriteRequestBuilder struct{
	simplePostCreateRequest
	simplePostUpdateRequest
}

var _ WriteRequestBuilder = PostPostWriteRequestBuilder{}

func (b PostPostWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type simplePostCreateRequest struct{}

func (simplePostCreateRequest) MakeCreateRequest(
	objectName string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header) {
	return clients.JSON.Post, nil
}

type simplePutUpdateRequest struct{}

func (simplePutUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Put, nil
}

type simpleNoopUpdateRequest struct{}

func (simpleNoopUpdateRequest) MakeUpdateRequest(
	string, string, *urlbuilder.URL, Clients) (common.WriteMethod, []common.Header) {
	return nil, nil
}

type simplePatchUpdateRequest struct{}

func (simplePatchUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Patch, nil
}

type simplePostUpdateRequest struct{}

func (simplePostUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Post, nil
}
