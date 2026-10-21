package orb

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/orb/internal/metadata"
)

const (
	// Spec paths are relative to the server https://api.withorb.com/v1.
	apiVersion = "v1"

	// Default page size is 20, the maximum 100.
	// https://docs.withorb.com/api-reference/pagination#limits
	maxPageSize = 100
)

// List invoices returns only issued, paid and synced invoices unless asked for others.
// Every status is requested, so that drafts are read and an invoice that is voided
// stays visible with its new status instead of dropping out of the read.
// https://docs.withorb.com/api-reference/invoice/list-invoices
var invoiceStatuses = []string{"draft", "issued", "paid", "synced", "void"} // nolint:gochecknoglobals

// buildReadRequest lists every record of the object, newest first, with cursor pagination.
// Every read is a full read: none of the list endpoints filters by update time and the records
// carry no update timestamp to filter on locally, while customers, subscriptions, invoices and
// metrics all change after they are created (status, balances, dates). Their created_at and
// invoice_date filters would miss those changes, so Since and Until are not sent.
// https://docs.withorb.com/api-reference/pagination
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	readURL, err := c.readURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

func (c *Connector) readURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam("limit", strconv.Itoa(min(readhelper.PageSizeWithDefault(params, maxPageSize), maxPageSize)))

	if params.ObjectName == objectNameInvoices {
		readURL.WithQueryParamList("status[]", invoiceStatuses)
	}

	return readURL, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(response,
		common.ExtractRecordsFromPath("data"),
		makeNextPageFunc(request.URL),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}
