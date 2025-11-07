package salesforce

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestBatchCreate(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errBadFirstRecord := testutils.DataFromFile(t, "batch/create/contacts/err-bad-request-on-first.json")
	errBadSecondRecord := testutils.DataFromFile(t, "batch/create/contacts/err-bad-request-on-second.json")
	createPayload := testutils.DataFromFile(t, "batch/create/contacts/payload.json")
	responseCreateContacts := testutils.DataFromFile(t, "batch/create/contacts/success.json")

	type record = common.Record

	createRecords := []any{
		record{
			"FirstName": "Siena",
			"LastName":  "Dyer",
		},
		record{
			"FirstName": "Markus",
			"LastName":  "Blevins",
		},
	}

	tests := []testroutines.BatchWrite{
		{
			Name: "At least one object name must be queried",
			Input: &common.BatchWriteParam{
				ObjectName: "",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Batch write type is missing",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrUnknownBatchWriteType},
		},
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeCreate,
				Records:    createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/tree/Contact"),
				Then:  mockserver.Response(http.StatusBadRequest, errBadFirstRecord),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: nil,
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
						  "statusCode": "REQUIRED_FIELD_MISSING",
						  "message": "Required fields are missing: [LastName]",
						  "fields": ["LastName"]}`)},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{
						common.ErrBatchUnprocessedRecord,
						"record's referenceId is ref1",
					},
					Data: nil,
				}},
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Many errors not traceable to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeCreate,
				Records:    createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/tree/Contact"),
				Then:  mockserver.Response(http.StatusBadRequest, errBadSecondRecord),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: nil,
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{
						common.ErrBatchUnprocessedRecord,
						"record's referenceId is ref0",
					},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{
						mockutils.JSONErrorWrapper(`{
						  "statusCode": "INVALID_INPUT",
						  "message": "Duplicate ReferenceId provided in the request.",
						  "fields": []}`),
						mockutils.JSONErrorWrapper(`{
						  "statusCode": "PROCESSING_HALTED",
						  "message": "Duplicate ReferenceId found: ref1",
						  "fields": []}`),
					},
					Data: nil,
				}},
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Bad request without the body",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeCreate,
				Records:    createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/tree/Contact"),
				Then:  mockserver.Response(http.StatusBadRequest, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{
						common.ErrBatchUnprocessedRecord,
						"record's referenceId is ref0",
					},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{
						common.ErrBatchUnprocessedRecord,
						"record's referenceId is ref1",
					},
					Data: nil,
				}},
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write with valid payload construction",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeCreate,
				Records:    createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/composite/tree/Contact"),
					mockcond.BodyBytes(createPayload), // validate that connector knows how to create payload.
				},
				Then: mockserver.Response(http.StatusOK, responseCreateContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusSuccess,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "003ak00000kLkZLAA0",
					Errors:   nil,
					Data: map[string]any{
						"referenceId": "ref0",
						"id":          "003ak00000kLkZLAA0",
					},
				}, {
					Success:  true,
					RecordId: "003ak00000kLkZMAA0",
					Errors:   nil,
					Data: map[string]any{
						"referenceId": "ref1",
						"id":          "003ak00000kLkZMAA0",
					},
				}},
				SuccessCount: 2,
				FailureCount: 0,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.BatchWriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestBatchUpdate(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errNoIDs := testutils.DataFromFile(t, "batch/update/contacts/err-each-no-ids.json")
	errUpdatePartial := testutils.DataFromFile(t, "batch/update/contacts/partial-no-ids.json")
	responseUpdateContacts := testutils.DataFromFile(t, "batch/update/contacts/success.json")

	type record = common.Record

	updateRecords := []any{
		record{
			"id":        "003ak00000jvIfpAAE",
			"FirstName": "Siena",
			"LastName":  "Dyer",
		},
		record{
			"id":        "003ak00000jvIfqAAE",
			"FirstName": "Markus",
			"LastName":  "Blevins",
		},
	}

	tests := []testroutines.BatchWrite{
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeUpdate,
				Records:    updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/sobjects"),
				Then:  mockserver.Response(http.StatusBadRequest, errNoIDs),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: nil,
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
							"statusCode": "MISSING_ARGUMENT",
							"message": "Id not specified in an update call",
							"fields": []}`)},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
							"statusCode": "MISSING_ARGUMENT",
							"message": "Id not specified in an update call",
							"fields": []}`)},
					Data: nil,
				}},
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Partial result where one contact did not have an id",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeUpdate,
				Records:    updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/sobjects"),
				Then:  mockserver.Response(http.StatusMultiStatus, errUpdatePartial),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusPartial,
				Errors: nil,
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "003ak00000jvIfpAAE",
					Errors:   nil,
					Data: map[string]any{
						"id":      "003ak00000jvIfpAAE",
						"success": true,
						"errors":  []any{},
					},
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
							"statusCode": "MISSING_ARGUMENT",
							"message": "Id not specified in an update call",
							"fields": []}`)},
					Data: nil,
				}},
				SuccessCount: 1,
				FailureCount: 1,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.BatchWriteTypeUpdate,
				Records:    updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/services/data/v60.0/composite/sobjects"),
				Then:  mockserver.Response(http.StatusOK, responseUpdateContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusSuccess,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "003ak00000jvIfpAAE",
					Errors:   nil,
					Data: map[string]any{
						"id":      "003ak00000jvIfpAAE",
						"success": true,
						"errors":  []any{},
					},
				}, {
					Success:  true,
					RecordId: "003ak00000jvIfqAAE",
					Errors:   nil,
					Data: map[string]any{
						"id":      "003ak00000jvIfqAAE",
						"success": true,
						"errors":  []any{},
					},
				}},
				SuccessCount: 2,
				FailureCount: 0,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.BatchWriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
