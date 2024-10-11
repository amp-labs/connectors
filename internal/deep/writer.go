package deep

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type Writer struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	support           dpobjects.Support
	urlResolver       dpobjects.URLResolver
	requester         dpwrite.Requester
	responder         dpwrite.Responder
}

func newWriter(
	clients *dprequests.Clients,
	headerSupplements *dprequests.HeaderSupplements,
	support dpobjects.Support,
	urlResolver dpobjects.URLResolver,
	requester dpwrite.Requester,
	responder dpwrite.Responder,
) *Writer {
	return &Writer{
		clients:           *clients,
		headerSupplements: *headerSupplements,
		support:           support,
		urlResolver:       urlResolver,
		requester:         requester,
		responder:         responder,
	}
}

func (w Writer) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !w.support.IsWriteSupported(config.ObjectName) {
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
		write, headers = w.requester.MakeCreateRequest(config.ObjectName, url, w.clients)
		headers = append(headers, w.headerSupplements.CreateHeaders()...)
	} else {
		write, headers = w.requester.MakeUpdateRequest(config.ObjectName, config.RecordId, url, w.clients)
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
	return w.responder.CreateWriteResult(config, body)
}

func (w Writer) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.Writer,
		Constructor: newWriter,
	}
}
