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
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var apiVersion = "v3" //nolint:gochecknoglobals

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

func (c *Connector) parseWriteResponse( //nolint:funlen
	ctx context.Context,
	params common.WriteParams,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := response.Body()
	if !ok {
		// Handle empty response
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordIDPaths := map[string]string{ //nolint:lll
		"smtp/email":                           "messageId",
		"transactionalSMS/sms":                 "messageId",
		"whatsapp/sendMessage":                 "messageId",
		"smtp/templates":                       "id",
		"emailCampaigns":                       "id",
		"emailCampaigns/images":                "url",
		"smsCampaigns":                         "id",
		"whatsappCampaigns":                    "id",
		"whatsappCampaigns/template":           "id",
		"contacts":                             "id",
		"contacts/folders":                     "id",
		"contacts/export":                      "processId",
		"contacts/import":                      "processId",
		"senders":                              "id",
		"senders/domains":                      "id",
		"webhooks":                             "id",
		"webhooks/export":                      "processId",
		"corporate/subAccount":                 "id",
		"corporate/ssoToken":                   "token",
		"corporate/subAccount/ssoToken":        "token",
		"corporate/subAccount/key":             "key",
		"corporate/group":                      "id",
		"corporate/user/invitation/send":       "id",
		"organization/user/invitation/send":    "invoice_id",
		"organization/user/update/permissions": "invoice_id",
		"feeds":                                "id",
		"companies":                            "id",
		"events":                               "id",
		"crm/attributes":                       "id",
		"companies/import":                     "processId",
		"crm/deals":                            "id",
		"crm/deals/import":                     "processId",
		"crm/tasks":                            "id",
		"crm/notes":                            "id",
		"conversations/messages":               "id",
		"conversations/pushedMessages":         "id",
		"orders/status/batch":                  "batchId",
		"categories":                           "id",
		"products":                             "id",
		"couponCollections":                    "id",
		"payments/requests":                    "id",
		"loyalty/config/programs":              "id",
		"ecommerce/activate":                   "id",
	}

	idPath, valid := recordIDPaths[params.ObjectName]
	if !valid {
		return &common.WriteResult{
			Success: true,
			Errors:  nil,
			Data:    nil,
		}, nil
	}

	// Try string first
	rawID, err := jsonquery.New(node).StrWithDefault(idPath, "id")

	if err != nil {
		// Try integer
		IntID, err := jsonquery.New(node).IntegerOptional(idPath)
		if err != nil {
			return nil, err
		}

		str := strconv.FormatInt(*IntID, 10)
		rawID = str
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: rawID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
