package deep

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

// Reader is a major connector component which provides Read functionality.
// Embed this into connector struct.
// Provide dpobjects.URLResolver into deep.Connector.
type Reader struct {
	clients           dprequests.Clients
	headerSupplements dprequests.HeaderSupplements
	objectSupport     dpobjects.Support
	urlResolver       dpobjects.URLResolver
	firstPage         dpread.PaginationStart
	nextPage          dpread.PaginationStep
	requester         dpread.Requester
	responder         dpread.Responder
}

func newReader(
	clients *dprequests.Clients,
	headerSupplements *dprequests.HeaderSupplements,
	objectManager dpobjects.Support,
	resolver dpobjects.URLResolver,
	firstPage dpread.PaginationStart,
	nextPage dpread.PaginationStep,
	requester dpread.Requester,
	responder dpread.Responder,
) *Reader {
	return &Reader{
		clients:           *clients,
		headerSupplements: *headerSupplements,
		objectSupport:     objectManager,
		urlResolver:       resolver,
		firstPage:         firstPage,
		nextPage:          nextPage,
		requester:         requester,
		responder:         responder,
	}
}

func (reader Reader) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !reader.objectSupport.IsReadSupported(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := reader.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	read, headers := reader.requester.MakeReadRequest(config.ObjectName, reader.clients)
	headers = append(headers, reader.headerSupplements.ReadHeaders()...)

	rsp, err := read(ctx, url, nil, headers...)
	if err != nil {
		return nil, err
	}

	recordsFunc, err := reader.responder.GetRecordsFunc(config)
	if err != nil {
		return nil, err
	}

	nextPageFunc := func(node *ajson.Node) (string, error) {
		return reader.nextPage.NextPage(config, url, node)
	}

	return common.ParseResult(
		rsp,
		recordsFunc,
		nextPageFunc,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (reader Reader) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := reader.urlResolver.FindURL(dpobjects.ReadMethod, reader.clients.BaseURL(), config.ObjectName)
	if err != nil {
		return nil, err
	}

	return reader.firstPage.FirstPage(config, url)
}

func (reader Reader) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.Reader,
		Constructor: newReader,
	}
}
