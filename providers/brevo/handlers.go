package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/brevo/metadata"
)

var apiVersion = "v3" //nolint:gochecknoglobals

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	path, err := metadata.Schemas.LookupURLPath(c.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	// first page pagination
	if supportLimitAndOffset.Has(params.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(pageSize))
		url.WithQueryParam("offset", "0")
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)
	requestURL := request.URL

	return common.ParseResult(
		response,
		common.GetRecordsUnderJSONPath(responseFieldName),
		nextRecordsURL(requestURL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		// Handle empty response
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordIDPaths := datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
		"smtp/email":                           "messageId",
		"transactionalSMS/sms":                 "messageId",
		"whatsapp/sendMessage":                 "messageId",
		"contacts/export":                      "processId",
		"contacts/import":                      "processId",
		"webhooks/export":                      "processId",
		"corporate/ssoToken":                   "token",
		"corporate/subAccount/ssoToken":        "token",
		"corporate/subAccount/key":             "key",
		"organization/user/invitation/send":    "invoice_id",
		"organization/user/update/permissions": "invoice_id",
		"companies/import":                     "processId",
		"crm/deals/import":                     "processId",
		"orders/status/batch":                  "batchId",
	},
		func(objectName string) string {
			return "id"
		},
	)

	idPath := recordIDPaths.Get(params.ObjectName)

	rawID, err := jsonquery.New(node).TextWithDefault(idPath, "")
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: rawID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
