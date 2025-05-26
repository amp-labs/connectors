package freshdesk

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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	contact := testutils.DataFromFile(t, "create-contacts.json")
	ticket := testutils.DataFromFile(t, "update-ticket.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Unsupported object name",
			Input: common.WriteParams{ObjectName: "butterflies", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Creation of a Contact",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "realdata"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/contacts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, contact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "501002381517",
				Data: map[string]any{
					"active":         false,
					"deleted":        false,
					"email":          "superman@freshdesk.com",
					"id":             float64(501002381517),
					"name":           "Super Man",
					"time_zone":      "Eastern Time (US & Canada)",
					"created_at":     "2025-02-17T11:05:13Z",
					"updated_at":     "2025-02-17T11:05:13Z",
					"first_name":     "Super",
					"last_name":      "Man",
					"visitor_id":     "df69c082-f2e6-4874-ad05-2ecfc7e27621",
					"org_contact_id": float64(1891443779623272448),
				},
			},
		},
		{
			Name: "Update of a ticket via PUT",
			Input: common.WriteParams{
				ObjectName: "tickets",
				RecordId:   "3",
				RecordData: "somenewdata",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/tickets/3"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, ticket),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "3",
				Data: map[string]any{
					"cc_emails":        []any{"ram@freshdesk.com", "diana@freshdesk.com"},
					"reply_cc_emails":  []any{"ram@freshdesk.com", "diana@freshdesk.com"},
					"ticket_cc_emails": []any{"ram@freshdesk.com", "diana@freshdesk.com"},
					"spam":             false,
					"email_config_id":  float64(501000047756),
					"fr_escalated":     false,
					"priority":         float64(1),
					"requester_id":     float64(501002381489),
					"source":           float64(2),
					"status":           float64(2),
					"subject":          "Support Needed...ASAP",
					"description":      "<div>Details about the issue...</div>",
					"description_text": "Details about the issue...",
					"id":               float64(3),
					"product_id":       float64(501000043986),
					"is_escalated":     false,
					"nr_escalated":     false,
					"created_at":       "2025-02-17T11:00:26Z",
					"updated_at":       "2025-02-17T11:15:00Z",
					"due_by":           "2025-02-19T22:00:00Z",
					"fr_due_by":        "2025-02-17T22:00:00Z",
					"form_id":          float64(501000123310),
					"sentiment_score":  float64(44),
				},
			},
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
