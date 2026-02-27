package salesforce

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

	responseUnknownField := testutils.DataFromFile(t, "unknown-field.json")
	responseInvalidFieldUpsert := testutils.DataFromFile(t, "invalid-field-upsert.json")
	responseCreateOK := testutils.DataFromFile(t, "create-ok.json")
	responseOKWithErrors := testutils.DataFromFile(t, "success-with-errors.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "account"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Error response understood for creating with unknown field",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseUnknownField),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("No such column 'AccountNumer' on sobject of type Account"),
			},
		},
		{
			Name:  "Error response understood for updating reserved field",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidFieldUpsert),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Unable to create/update fields: MasterRecordId"),
			},
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "account", RecordId: "003ak000004dQCUAA2", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.QueryParam("_HttpMethod", "PATCH"),
				},
				Then: mockserver.Response(http.StatusOK, nil), // real salesforce returns an empty body
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "003ak000004dQCUAA2",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of account",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateOK),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "001ak00000OQTieAAH",
				Errors:   []any{},
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "OK Response, but with errors field",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseOKWithErrors),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  false,
				RecordId: "001RM000003oLruYAE",
				Errors: []any{map[string]any{
					"statusCode": "MALFORMED_ID",
					"message":    "malformed id 001RM000003oLrB000",
					"fields":     []any{},
				}},
				Data: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestWritePardot(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseCreateOK := testutils.DataFromFile(t, "pardot/write/prospect/new.json")

	pardotHeader := http.Header{
		"Pardot-Business-Unit-Id": []string{"test-business-unit-id"},
	}

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "prospects"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Create a prospect",
			Input: common.WriteParams{ObjectName: "proSPEcTs", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v5/objects/prospects"),
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateOK),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "55434583",
				Errors:   nil,
				Data: map[string]any{
					"id":        float64(55434583),
					"email":     "a.alexander@sample.com",
					"firstName": "Athenasius",
					"lastName":  "Alexander",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update a prospect",
			Input: common.WriteParams{ObjectName: "prospects", RecordId: "55434583", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/api/v5/objects/prospects/55434583"),
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateOK),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "55434583",
				Errors:   nil,
				Data: map[string]any{
					"id":        float64(55434583),
					"email":     "a.alexander@sample.com",
					"firstName": "Athenasius",
					"lastName":  "Alexander",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnectorAccountEngagement(tt.Server.URL)
			})
		})
	}
}
