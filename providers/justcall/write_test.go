package justcall

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	createContactResponse := testutils.DataFromFile(t, "write/contacts/create.json")
	updateContactResponse := testutils.DataFromFile(t, "write/contacts/update.json")
	sendSMSResponse := testutils.DataFromFile(t, "write/texts/send.json")
	updateCallResponse := testutils.DataFromFile(t, "write/calls/update.json")
	successResponse := testutils.DataFromFile(t, "write/success.json")
	errorResponse := testutils.DataFromFile(t, "write/error.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error on bad request",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name:  "Create contact",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/contacts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createContactResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"id":     float64(12345),
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "12345",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/contacts"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, updateContactResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"id":     float64(12345),
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Send SMS via texts object",
			Input: common.WriteParams{ObjectName: "texts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/texts/new"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, sendSMSResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "98765",
				Data: map[string]any{
					"id":     float64(98765),
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update call with path ID",
			Input: common.WriteParams{
				ObjectName: "calls",
				RecordId:   "328951212",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/calls/328951212"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, updateCallResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "328951212",
				Data: map[string]any{
					"id":     float64(328951212),
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update contact status uses PUT without RecordId",
			Input: common.WriteParams{ObjectName: "contacts/status", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/contacts/status"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Add tag to thread",
			Input: common.WriteParams{ObjectName: "texts/threads/tag", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/texts/threads/tag"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Add contact to campaign",
			Input: common.WriteParams{ObjectName: "sales_dialer/campaigns/contact", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/sales_dialer/campaigns/contact"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Initiate voice agent call",
			Input: common.WriteParams{ObjectName: "voice-agents/calls", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2.1/voice-agents/calls"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"status": "success",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
