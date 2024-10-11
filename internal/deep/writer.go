package deep

import (
	"context"
	"errors"
	"github.com/amp-labs/connectors/internal/deep/requirements"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
)

type Writer struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	objectManager     dpobjects.ObjectManager
	urlResolver       dpobjects.ObjectURLResolver
	requestBuilder    dpwrite.WriteRequestBuilder
	resultBuilder     dpwrite.WriteResultBuilder
}

func newWriter(clients *dprequests.Clients,
	resolver dpobjects.ObjectURLResolver,
	requestBuilder dpwrite.WriteRequestBuilder,
	resultBuilder *dpwrite.WriteResultBuilder,
	objectManager dpobjects.ObjectManager,
	headerSupplements *dprequests.HeaderSupplements,
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

func (w Writer) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
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

func (w Writer) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.Writer,
		Constructor: newWriter,
	}
}
