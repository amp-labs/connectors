package qargo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/qargo/internal/metadata"
)

const (
	// https://api-docs.qargo.com/docs/section/api-best-practices/pagination
	cursorParam = "cursor"

	// "Filter to fetch the companies updated after this time. Format is ISO 8601 YYYY-MM-DDTHH:MM:SSZ"
	// https://api-docs.qargo.com/docs/api-company/list_companies_v1_companies_company_get
	updatedAfterParam = "updated_after"
)

// buildReadRequest lists companies, or the sync-task queue for invoices and credit notes (queue.go),
// which takes no parameters.
// For companies the first page carries updated_after; later pages send only the cursor, as the docs
// require ("Do not combine cursor with other query parameters"), and the cursor keeps the filter of
// the request that started it.
// There is no page size parameter and no updated-before filter, so Until is applied in parseReadResponse.
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if _, isQueue := queueTaskTypes[params.ObjectName]; isQueue {
		return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
	}

	if params.NextPage != "" {
		readURL.WithQueryParam(cursorParam, params.NextPage.String())
	} else if !params.Since.IsZero() {
		// updated_after is exclusive and takes whole seconds, so it starts a second early to keep a
		// record updated exactly at Since; parseReadResponse trims the extra second.
		readURL.WithQueryParam(updatedAfterParam, datautils.Time.FormatRFC3339inUTC(params.Since.Add(-time.Second)))
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if _, isQueue := queueTaskTypes[params.ObjectName]; isQueue {
		return c.parseQueueResponse(ctx, params, response)
	}

	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc("items"),
		makeUpdatedFilter(),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}

// buildWriteRequest reports the payment status of a sales invoice. RecordId is the invoice id and
// RecordData is the documented body, {"status": "PAID" | "UNPAID"}.
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if params.IsCreate() {
		return nil, fmt.Errorf("%w: %s is update only, the record id is the sales invoice id",
			common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	writeURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion,
		"accounting/sales-invoice", params.RecordId, "payment/update-status")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, writeURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// parseWriteResponse returns the invoice id the status was reported for; the response,
// {"status": "PAID"}, carries no id of its own.
func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: params.RecordId,
		Data:     data,
	}, nil
}
