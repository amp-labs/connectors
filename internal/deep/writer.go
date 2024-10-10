package deep

import (
	"context"
	"errors"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

type Writer struct {
	urlResolver   dpobjects.ObjectURLResolver
	resultBuilder  WriteResultBuilder
	objectManager  dpobjects.ObjectManager
	requestBuilder WriteRequestBuilder
	headerSupplements HeaderSupplements

	clients Clients
}

func NewWriter(clients *Clients,
	resolver dpobjects.ObjectURLResolver,
	requestBuilder WriteRequestBuilder,
	resultBuilder *WriteResultBuilder,
	objectManager dpobjects.ObjectManager,
	headerSupplements *HeaderSupplements,
) *Writer {
	return &Writer{
		urlResolver:       resolver,
		resultBuilder:     *resultBuilder,
		objectManager:     objectManager,
		requestBuilder:    requestBuilder,
		headerSupplements: *headerSupplements,
		clients:           *clients,
	}
}

func (w *Writer) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !w.objectManager.IsWriteSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	method := dpobjects.CreateMethod
	if len(config.RecordId) != 0 {
		method = dpobjects.UpdateMethod
	}

	url, err := w.urlResolver.FindURL(method, w.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	var headers []common.Header
	if len(config.RecordId) == 0 {
		write, headers = w.requestBuilder.MakeCreateRequest(config.ObjectName, url, w.clients)
		headers = append(headers, w.headerSupplements.CreateHeaders()...)
	} else {
		write, headers = w.requestBuilder.MakeUpdateRequest(config.ObjectName, config.RecordId, url, w.clients)
		if write == nil {
			// TODO need a better error
			return nil, errors.New("update is not supported for this object")
		}

		headers = append(headers, w.headerSupplements.UpdateHeaders()...)
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
	requirements.ConnectorComponent

	MakeCreateRequest(
		objectName string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header)
	MakeUpdateRequest(
		objectName, recordID string, url *urlbuilder.URL, clients Clients) (common.WriteMethod, []common.Header)
}

type PostPutWriteRequestBuilder struct {
	SimplePostCreateRequest
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

type PostWriteRequestBuilder struct {
	SimplePostCreateRequest
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

type PostPatchWriteRequestBuilder struct {
	SimplePostCreateRequest
	SimplePatchUpdateRequest
}

var _ WriteRequestBuilder = PostPatchWriteRequestBuilder{}

func (b PostPatchWriteRequestBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
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
		ID:          "writeRequestBuilder",
		Constructor: handy.Returner(b),
		Interface:   new(WriteRequestBuilder),
	}
}

type SimplePostCreateRequest struct{}

func (SimplePostCreateRequest) MakeCreateRequest(
	objectName string, url *urlbuilder.URL, clients Clients,
) (common.WriteMethod, []common.Header) {
	return clients.JSON.Post, nil
}

type simplePutUpdateRequest struct{}

func (simplePutUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Put, nil
}

type simpleNoopUpdateRequest struct{}

func (simpleNoopUpdateRequest) MakeUpdateRequest(
	string, string, *urlbuilder.URL, Clients,
) (common.WriteMethod, []common.Header) {
	return nil, nil
}

type SimplePatchUpdateRequest struct{}

func (SimplePatchUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Patch, nil
}

type SimplePostUpdateRequest struct{}

func (SimplePostUpdateRequest) MakeUpdateRequest(
	objectName string, recordID string, url *urlbuilder.URL, clients Clients,
) (common.WriteMethod, []common.Header) {
	url.AddPath(recordID)

	return clients.JSON.Post, nil
}
