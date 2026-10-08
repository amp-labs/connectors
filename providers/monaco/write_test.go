package monaco

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseContactCreate := testutils.DataFromFile(t, "write/contact-create.json")
	responseContactUpdate := testutils.DataFromFile(t, "write/contact-update.json")
	responseAccountUpsert := testutils.DataFromFile(t, "write/account-upsert.json")
	responseSequenceTemplate := testutils.DataFromFile(t, "write/sequence-template-create.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs record data",
			Input:        common.WriteParams{ObjectName: objectContacts},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Unsupported object is rejected",
			Input: common.WriteParams{
				ObjectName: objectUsers,
				RecordData: map[string]any{"email": "x@y.com"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Meetings are read-only",
			Input: common.WriteParams{
				ObjectName: objectMeetings,
				RecordData: map[string]any{"title": "Kickoff"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Sequences cannot be created, only updated",
			Input: common.WriteParams{
				ObjectName: objectSequences,
				RecordData: map[string]any{"action": "start"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Audiences cannot be updated, only created",
			Input: common.WriteParams{
				ObjectName: objectAudiences,
				RecordId:   "aud_001",
				RecordData: map[string]any{"name": "Renamed"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create contact POSTs to the slash-terminated collection",
			Input: common.WriteParams{
				ObjectName: objectContacts,
				RecordData: map[string]any{
					"email":      "grace@acme.com",
					"first_name": "Grace",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					// /v1/contacts would answer 307.
					mockcond.Path("/v1/contacts/"),
					mockcond.Body(`{"email":"grace@acme.com","first_name":"Grace"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "con_new001",
				Errors:   nil,
				Data: map[string]any{
					"id":    "con_new001",
					"email": "grace@acme.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact PATCHes the record path",
			Input: common.WriteParams{
				ObjectName: objectContacts,
				RecordId:   "con_abc123",
				RecordData: map[string]any{"title": "CTO"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1/contacts/con_abc123"),
					mockcond.Body(`{"title":"CTO"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "con_abc123",
				Errors:   nil,
				Data:     map[string]any{"title": "CTO"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Server-assigned id is stripped from the request body",
			Input: common.WriteParams{
				ObjectName: objectContacts,
				RecordId:   "con_abc123",
				RecordData: map[string]any{"id": "con_abc123", "title": "CTO"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					// No "id" key -- it appears in no Monaco request schema.
					mockcond.Body(`{"title":"CTO"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "con_abc123",
				Errors:   nil,
				Data:     map[string]any{"title": "CTO"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create account falls back to PUT upsert",
			Input: common.WriteParams{
				ObjectName: objectAccounts,
				RecordData: map[string]any{
					"domain": "acme.com",
					"name":   "Acme Corp",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					// POST /v1/accounts/ answers 405; PUT is the only insert path.
					mockcond.MethodPUT(),
					mockcond.Path("/v1/accounts/"),
					mockcond.Body(`{"domain":"acme.com","name":"Acme Corp"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountUpsert),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_def456",
				Errors:   nil,
				Data:     map[string]any{"name": "Acme Corp"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update account still uses PATCH",
			Input: common.WriteParams{
				ObjectName: objectAccounts,
				RecordId:   "acc_def456",
				RecordData: map[string]any{"status": "active"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1/accounts/acc_def456"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountUpsert),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_def456",
				Errors:   nil,
				Data:     map[string]any{"status": "active"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Sequence templates create on the unslashed route and answer 201",
			Input: common.WriteParams{
				ObjectName: objectSequenceTemplates,
				RecordData: map[string]any{"name": "Outbound v2"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					// Mirror image of contacts: the slashed form is the redirect.
					mockcond.Path("/v1/sequence-templates"),
				},
				Then: mockserver.Response(http.StatusCreated, responseSequenceTemplate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "seqt_001",
				Errors:   nil,
				Data:     map[string]any{"name": "Outbound v2"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Audiences create on the unslashed route",
			Input: common.WriteParams{
				ObjectName: objectAudiences,
				RecordData: map[string]any{"name": "Q3 targets"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/audiences"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"data":{"id":"aud_001","name":"Q3 targets"}}`)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "aud_001",
				Errors:   nil,
				Data:     map[string]any{"name": "Q3 targets"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Response without a data envelope still reports success",
			Input: common.WriteParams{
				ObjectName: objectTags,
				RecordId:   "tag_001",
				RecordData: map[string]any{"name": "Hot"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, []byte(`{"meta":{}}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "tag_001",
			},
			ExpectedErrs: nil,
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
