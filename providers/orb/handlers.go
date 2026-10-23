package orb

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/orb/internal/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion = "v1"

	// List endpoints default to 20 items and accept at most 100.
	// https://docs.withorb.com/api-reference/pagination
	maxPageSize = 100
)

// List invoices returns only `issued`, `paid` and `synced` invoices unless `status[]` is given,
// so every status is requested: a full read must also see invoices that were voided after they
// were issued, and drafts. Draft values are Orb's last computation and may lag behind its refresh.
// https://docs.withorb.com/api-reference/invoice/list-invoices
var invoiceStatuses = []string{"draft", "issued", "paid", "synced", "void"} // nolint:gochecknoglobals

// buildReadRequest lists an object with cursor pagination.
//
// Every object is read in full: Since and Until are not sent. Orb's list endpoints filter only by
// `created_at` (customers, subscriptions, metrics) or by invoice and due date (invoices), and none
// of the records carries an updated-at timestamp, so neither the provider nor the connector can
// tell which records changed. A creation-time filter would miss updates such as a subscription
// ending or an invoice being paid.
// https://docs.withorb.com/api-reference/subscription/list-subscriptions
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
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

	if params.NextPage != "" {
		readURL.WithQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(response,
		common.MakeRecordsFunc("data"),
		nextPageCursor,
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// nextPageCursor returns `pagination_metadata.next_cursor` while `has_more` is true.
// https://docs.withorb.com/api-reference/pagination
func nextPageCursor(node *ajson.Node) (string, error) {
	pagination := jsonquery.New(node, "pagination_metadata")

	hasMore, err := pagination.BoolWithDefault("has_more", false)
	if err != nil || !hasMore {
		return "", err
	}

	return pagination.StrWithDefault("next_cursor", "")
}
