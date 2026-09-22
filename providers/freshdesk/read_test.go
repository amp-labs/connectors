package freshdesk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "empty.json")
	mailboxes := testutils.DataFromFile(t, "mailboxes.json")
	tickets := testutils.DataFromFile(t, "tickets.json")
	settings := testutils.DataFromFile(t, "settings.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "meme", Fields: datautils.NewStringSet("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Mailboxes",
			Input: common.ReadParams{
				ObjectName: "mailboxes",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/email/mailboxes"),
				Then:  mockserver.Response(http.StatusOK, mailboxes),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(1),
						"name": "Test Account",
					},
					Raw: map[string]any{
						"id":                  float64(1),
						"name":                "Test Account",
						"support_email":       "testemail@gmail.com",
						"group_id":            float64(1),
						"default_reply_email": true,
						"active":              true,
						"mailbox_type":        "freshdesk_mailbox",
						"created_at":          "2019-08-20T05:55:38Z",
						"updated_at":          "2019-08-20T05:55:38Z",
						"product_id":          float64(2),
						"freshdesk_mailbox": map[string]any{
							"forward_email": "gmailcomtestemail@freshdesk.com",
						},
					},
					Id: "1",
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			// Settings answers with a lone object, which stands for a single record.
			Name:  "Singleton object is read as one record",
			Input: common.ReadParams{ObjectName: "settings", Fields: connectors.Fields("primary_language")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/settings/helpdesk"),
				Then:  mockserver.Response(http.StatusOK, settings),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"primary_language": "en",
					},
					Raw: map[string]any{
						"primary_language":      "en",
						"supported_languages":   []any{},
						"portal_languages":      []any{},
						"help_widget_languages": []any{},
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			// Freshdesk points at the page that follows through the Link header.
			Name:  "Next page is read from the Link header",
			Input: common.ReadParams{ObjectName: "tickets", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/tickets"),
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://ampersand.freshdesk.com/api/v2/tickets?per_page=1&page=2>; rel="next"`),
					mockserver.Response(http.StatusOK, tickets),
				),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "https://ampersand.freshdesk.com/api/v2/tickets?per_page=1&page=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			// The header is absent on the last page, which is how a read ends.
			Name:  "Read is done when no Link header is sent",
			Input: common.ReadParams{ObjectName: "tickets", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/tickets"),
				Then:  mockserver.Response(http.StatusOK, tickets),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
