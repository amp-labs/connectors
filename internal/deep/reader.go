package deep

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
)

type Reader struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	objectManager     dpobjects.ObjectManager
	urlResolver       dpobjects.ObjectURLResolver
	pageStartBuilder  dpread.PaginationStartBuilder
	nextPageBuilder   dpread.NextPageBuilder
	readObjectLocator dpread.ReadObjectLocator
	requestBuilder    dpread.ReadRequestBuilder
}

func NewReader(clients *dprequests.Clients,
	resolver dpobjects.ObjectURLResolver,
	pageStartBuilder dpread.PaginationStartBuilder,
	nextPageBuilder *dpread.NextPageBuilder,
	objectLocator *dpread.ReadObjectLocator,
	objectManager dpobjects.ObjectManager,
	requestBuilder dpread.ReadRequestBuilder,
	headerSupplements *dprequests.HeaderSupplements,
) *Reader {
	return &Reader{
		urlResolver:       resolver,
		pageStartBuilder:  pageStartBuilder,
		nextPageBuilder:   *nextPageBuilder,
		readObjectLocator: *objectLocator,
		objectManager:     objectManager,
		requestBuilder:    requestBuilder,
		headerSupplements: *headerSupplements,
		clients:           *clients,
	}
}

func (r *Reader) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !r.objectManager.IsReadSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := r.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	read, headers := r.requestBuilder.MakeReadRequest(config.ObjectName, r.clients)
	headers = append(headers, r.headerSupplements.ReadHeaders()...)

	rsp, err := read(ctx, url, nil, headers...)
	if err != nil {
		return nil, err
	}

	recordsFunc, err := r.readObjectLocator.GetRecordsFunc(config)
	if err != nil {
		return nil, err
	}

	nextPageFunc, err := r.nextPageBuilder.GetNextPageFunc(config, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		recordsFunc,
		nextPageFunc,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (r *Reader) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := r.urlResolver.FindURL(dpobjects.ReadMethod, r.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	return r.pageStartBuilder.FirstPage(config, url)
}
