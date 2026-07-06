package acculynx

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

//go:embed test/write/contact-create.json
var contactCreateResponse []byte

//go:embed test/write/job-create.json
var jobCreateResponse []byte

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Object must be supported",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordData: map[string]any{"name": "test"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update on contacts is not supported",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "ctc_001",
				RecordData: map[string]any{"firstName": "Carol"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create on jobs/custom-fields is not supported",
			Input: common.WriteParams{
				ObjectName: "jobs/custom-fields",
				RecordData: map[string]any{"jobId": "job_001"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create contact succeeds and returns recordId",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName":      "Carol",
					"lastName":       "Customer",
					"contactTypeIds": []string{"52ba94c5-3ecf-4e7f-90cd-a91de12a72f5"},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/contacts"),
						},
						Then: mockserver.Response(http.StatusCreated, contactCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ctc_001",
				Data: map[string]any{
					"id": "ctc_001",
					// Raw response field (_link) must round-trip through Data,
					// proving the connector doesn't strip provider-returned fields.
					"_link": "https://api.acculynx.com/api/v2/contacts/ctc_001",
				},
			},
		},
		{
			Name: "Create job succeeds and returns recordId",
			Input: common.WriteParams{
				ObjectName: "jobs",
				RecordData: map[string]any{
					"contact": map[string]any{"id": "ctc_001"},
					"notes":   "New roof",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs"),
						},
						Then: mockserver.Response(http.StatusCreated, jobCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "job_001",
				Data: map[string]any{
					"id":    "job_001",
					"_link": "https://api.acculynx.com/api/v2/jobs/job_001",
				},
			},
		},
		{
			Name: "Update jobs/custom-fields succeeds with 204 and echoes recordId",
			Input: common.WriteParams{
				ObjectName: "jobs/custom-fields",
				RecordId:   "cf_001",
				RecordData: map[string]any{
					"jobId":     "job_001",
					"fieldType": "Text",
					"values":    []string{"Updated value"},
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/api/v2/jobs/job_001/custom-fields/cf_001"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cf_001",
			},
		},
		{
			Name: "Update jobs/initial-appointment routes to PUT and uses jobId from record",
			Input: common.WriteParams{
				ObjectName: "jobs/initial-appointment",
				RecordData: map[string]any{
					"jobId":     "job_001",
					"startDate": "2026-06-22T18:47:10Z",
					"endDate":   "2026-06-22T19:47:10Z",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/api/v2/jobs/job_001/initial-appointment"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "Update jobs/insurance/insurance-company routes to PUT",
			Input: common.WriteParams{
				ObjectName: "jobs/insurance/insurance-company",
				RecordData: map[string]any{
					"jobId":              "job_001",
					"insuranceCompanyId": "ins_001",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/api/v2/jobs/job_001/insurance/insurance-company"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/representatives/ar-owner routes correctly",
			Input: common.WriteParams{
				ObjectName: "jobs/representatives/ar-owner",
				RecordData: map[string]any{
					"jobId": "job_001",
					"id":    "user_001",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/representatives/ar-owner"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/representatives/sales-owner routes correctly",
			Input: common.WriteParams{
				ObjectName: "jobs/representatives/sales-owner",
				RecordData: map[string]any{
					"jobId": "job_001",
					"id":    "user_002",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/representatives/sales-owner"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/representatives/company routes correctly",
			Input: common.WriteParams{
				ObjectName: "jobs/representatives/company",
				RecordData: map[string]any{
					"jobId": "job_001",
					"id":    "user_003",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/representatives/company"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST contacts/phone-numbers fans out under the parent contact",
			Input: common.WriteParams{
				ObjectName: "contacts/phone-numbers",
				RecordData: map[string]any{
					"contactId": "ctc_001",
					"number":    "5551234567",
					"type":      "Mobile",
					"primary":   true,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/contacts/ctc_001/phone-numbers"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST contacts/logs fans out under the parent contact",
			Input: common.WriteParams{
				ObjectName: "contacts/logs",
				RecordData: map[string]any{
					"contactId": "ctc_001",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/contacts/ctc_001/logs"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/external-references routes to top-level path (no parent in URL)",
			Input: common.WriteParams{
				ObjectName: "jobs/external-references",
				RecordData: map[string]any{
					"jobId":  "job_001",
					"source": "test-source",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/external-references"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/messages fans out under parent job",
			Input: common.WriteParams{
				ObjectName: "jobs/messages",
				RecordData: map[string]any{
					"jobId":   "job_001",
					"message": "Hello world",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/messages"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/messages/replies uses two-level fan-out (jobId + messageId)",
			Input: common.WriteParams{
				ObjectName: "jobs/messages/replies",
				RecordData: map[string]any{
					"jobId":     "job_001",
					"messageId": "msg_001",
					"message":   "Reply text",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/messages/msg_001/replies"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/payments/expense fans out under parent job",
			Input: common.WriteParams{
				ObjectName: "jobs/payments/expense",
				RecordData: map[string]any{
					"jobId":  "job_001",
					"amount": 100,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/payments/expense"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/payments/paid fans out under parent job",
			Input: common.WriteParams{
				ObjectName: "jobs/payments/paid",
				RecordData: map[string]any{
					"jobId":  "job_001",
					"amount": 200,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/payments/paid"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
		{
			Name: "POST jobs/payments/received fans out under parent job",
			Input: common.WriteParams{
				ObjectName: "jobs/payments/received",
				RecordData: map[string]any{
					"jobId":  "job_001",
					"amount": 300,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/api/v2/jobs/job_001/payments/received"),
						},
						Then: mockserver.Response(http.StatusCreated, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected:   &common.WriteResult{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestWriteConnector(tt.Server)
			})
		})
	}
}

func constructTestWriteConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}
