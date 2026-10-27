package mail

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseDrafts := testutils.DataFromFile(t, "write/drafts/new.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "drafts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Create Draft",
			Input: common.WriteParams{ObjectName: "drafts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/gmail/v1/users/me/drafts"),
				},
				Then: mockserver.Response(http.StatusOK, responseDrafts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "r5151461288000502025",
				Errors:   nil,
				Data: map[string]any{
					"id": "r5151461288000502025",
					"message": map[string]any{
						"id":       "199400f21a8f1186",
						"threadId": "199400f21a8f1186",
						"labelIds": []any{
							"DRAFT",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update Draft",
			Input: common.WriteParams{
				ObjectName: "drafts",
				RecordData: "dummy",
				RecordId:   "r5151461288000502025",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/gmail/v1/users/me/drafts/r5151461288000502025"),
				},
				Then: mockserver.Response(http.StatusOK, responseDrafts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "r5151461288000502025",
				Errors:   nil,
				Data: map[string]any{
					"id": "r5151461288000502025",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:         "Creating messages is not supported",
			Input:        common.WriteParams{ObjectName: "messages", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Creating sent messages is not supported",
			Input:        common.WriteParams{ObjectName: "AMPERSAND-sentMessages", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Send message",
			Input: common.WriteParams{
				ObjectName: "AMPERSAND-messages",
				RecordData: map[string]string{"raw": "someBase64EncodedMessage"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/gmail/v1/users/me/messages/send"),
					mockcond.Body(`{"raw":"someBase64EncodedMessage"}`),
				},
				Then: mockserver.Response(http.StatusAccepted),
			}.Server(),
			Comparator:   testconn.ComparatorSubsetWrite,
			Expected:     &common.WriteResult{Success: true, RecordId: "", Data: nil},
			ExpectedErrs: nil,
		},
		{
			Name:         "Updating virtual messages is not valid",
			Input:        common.WriteParams{ObjectName: "AMPERSAND-messages", RecordData: "dummy", RecordId: "723"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestAdapter(tt.Server)
			})
		})
	}
}
