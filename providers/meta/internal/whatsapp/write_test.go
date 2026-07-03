package whatsapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const (
	testPhoneNumberID     = "123456789012345"
	testWhatsAppAccountID = "987654321098765"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	messageCreateResponse := testutils.DataFromFile(t, "message-create.json")
	messageTemplateCreateResponse := testutils.DataFromFile(t, "message-template-create.json")
	phoneNumberCreateResponse := testutils.DataFromFile(t, "phone-number-create.json")

	tests := []testroutines.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Object must be supported",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{"name": "test"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create text message uses phone number scoped URL",
			Input: common.WriteParams{
				ObjectName: "messages",
				RecordData: map[string]any{
					"messaging_product": "whatsapp",
					"recipient_type":    "individual",
					"to":                "+16505551234",
					"type":              "text",
					"text": map[string]any{
						"preview_url": true,
						"body": "As requested, here's the link to our latest product: " +
							"https://www.meta.com/quest/quest-3/",
					},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v25.0/" + testPhoneNumberID + "/messages"),
						},
						Then: mockserver.Response(http.StatusOK, messageCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "wamid.HBgLMTY0NjcwNDM1OTUVAgARGBI4MjZGRDA0OUE2OTQ3RkEyMzcA",
				Data: map[string]any{
					"messaging_product": "whatsapp",
					"contacts": []any{
						map[string]any{
							"input": "+16505551234",
							"wa_id": "16505551234",
						},
					},
				},
			},
		},
		{
			Name: "Create message template uses WABA scoped URL",
			Input: common.WriteParams{
				ObjectName: "message_templates",
				RecordData: map[string]any{
					"name": "seasonal_promotion",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v25.0/" + testWhatsAppAccountID + "/message_templates"),
						},
						Then: mockserver.Response(http.StatusOK, messageTemplateCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2450146205448663",
				Data: map[string]any{
					"id":       "2450146205448663",
					"status":   "PENDING",
					"category": "MARKETING",
				},
			},
		},
		{
			Name: "Create phone number uses WABA scoped URL",
			Input: common.WriteParams{
				ObjectName: "phone_numbers",
				RecordData: map[string]any{
					"cc":            "1",
					"phone_number":  "14195551518",
					"verified_name": "Lucky Shrub",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v25.0/" + testWhatsAppAccountID + "/phone_numbers"),
						},
						Then: mockserver.Response(http.StatusOK, phoneNumberCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1906385232743451",
				Data: map[string]any{
					"successful_creation": map[string]any{
						"summary": "Phone number successfully created",
						"value": map[string]any{
							"id": "1906385232743451",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableWriter, error) {
				return constructTestAdapter(tt.Server)
			})
		})
	}
}

func constructTestAdapter(server *httptest.Server) (*Adapter, error) {
	adapter, err := NewAdapter(common.ConnectorParams{
		Module:              providers.ModuleMetaWhatsApp,
		AuthenticatedClient: server.Client(),
		Metadata: map[string]string{
			"whatsappPhoneNumberId": testPhoneNumberID,
			"whatsappAccountId":     testWhatsAppAccountID,
		},
	})
	if err != nil {
		return nil, err
	}

	adapter.SetUnitTestBaseURL(server.URL)

	return adapter, nil
}
