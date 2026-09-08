package reply

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContactCreate := testutils.DataFromFile(t, "write/contact-create.json")
	responseContactUpdate := testutils.DataFromFile(t, "write/contact-update.json")
	responseCustomFieldUpdate := testutils.DataFromFile(t, "write/custom-field-update.json")
	responseSequenceFolderCreate := testutils.DataFromFile(t, "write/sequence-folder-create.json")
	errorMissingEmail := testutils.DataFromFile(t, "write/err-missing-email.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Read-only object is rejected by the registry",
			Input:        common.WriteParams{ObjectName: "inbox/threads", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Create contact via POST",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts"),
				},
				Then: mockserver.Response(http.StatusCreated, responseContactCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Errors:   nil,
				Data: map[string]any{
					"email":     "probe.write@example.com",
					"firstName": "Probe",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact via PATCH",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "761575530",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/761575530"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Errors:   nil,
				Data: map[string]any{
					"title": "Updated",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update custom field via PUT",
			Input: common.WriteParams{
				ObjectName: "custom-fields",
				RecordId:   "150697",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/custom-fields/150697"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomFieldUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "150697",
				Errors:   nil,
				Data: map[string]any{
					"title": "Deal Stage",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact customFields from a ReadResult round-trip: key becomes id",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"customFields": []any{map[string]any{
						"key":   "150697",
						"value": "probe",
					}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/759458702"),
					mockcond.Body(`{"customFields":[{"id":150697,"value":"probe"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data: map[string]any{
					"email": "probe.write@example.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create sequence folder returns a GUID record id",
			Input: common.WriteParams{ObjectName: "sequence-folders", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/sequence-folders"),
				},
				Then: mockserver.Response(http.StatusCreated, responseSequenceFolderCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "42b54f7a-b054-480b-ad1b-6f924c05ea4b",
				Errors:   nil,
				Data: map[string]any{
					"name": "Fixture Capture Folder",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Validation problem on create carries pointer detail",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorMissingEmail),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Validation failed: /firstName FirstName is required."),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
