package instantly

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseBlocklistEntry := testutils.DataFromFile(t, "write-blocklist-entry.json")

	responseReply := testutils.DataFromFile(t, "write-unibox-reply.json")

	responseLeadErr := testutils.DataFromFile(t, "write-lead-bad-request.json")
	responseLead := testutils.DataFromFile(t, "write-lead.json")

	responseTagErr := testutils.DataFromFile(t, "write-tag-bad-request.json")
	responseTag := testutils.DataFromFile(t, "write-tag.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "notes"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "unibox-replies", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "orders", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Create Blocklist Entry",
			Input: common.WriteParams{ObjectName: "blocklist-entries", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseBlocklistEntry),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cf8fd143-08c3-438e-a396-491aa1ced9d4",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create Unibox Reply",
			Input: common.WriteParams{ObjectName: "unibox-replies", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseReply),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "19e9d7f9-bd4f-45aa-8d77-9eecc9bd3f9a",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Invalid Lead creation",
			Input: common.WriteParams{ObjectName: "leads", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseLeadErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Leads array is empty"), // nolint:goerr113
			},
		},
		{
			Name:  "Create Lead",
			Input: common.WriteParams{ObjectName: "leads", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseLead),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "", // lead doesn't have ID
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Invalid Tag creation",
			Input: common.WriteParams{ObjectName: "tags", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseTagErr),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Bad Request"), // nolint:goerr113
			},
		},
		{
			Name:  "Create Tag acts as POST",
			Input: common.WriteParams{ObjectName: "tags", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseTag),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "f6825fcf-c51b-4724-937b-0814ed02af83",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update tag acts as PATCH",
			Input: common.WriteParams{ObjectName: "tags", RecordId: "885633", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, responseTag),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "f6825fcf-c51b-4724-937b-0814ed02af83",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
