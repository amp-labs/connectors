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

	errBadRequest := testutils.DataFromFile(t, "batch/create/contacts/err-bad-request.json")
	errCreatePartial := testutils.DataFromFile(t, "batch/create/contacts/partial-but-allOrNone.json")
	createPayload := testutils.DataFromFile(t, "batch/create/contacts/payload.json")
	responseCreateContacts := testutils.DataFromFile(t, "batch/create/contacts/success.json")

	type record = common.Record

	createRecords := common.BatchItems{{
		Record: record{
			"FirstName": "Siena",
			"LastName":  "Dyer",
		},
	}, {
		Record: record{
			"FirstName": "Markus",
			"LastName":  "Blevins",
		},
	}}

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
			ExpectedErrs: []error{common.ErrUnknownWriteType},
		},
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errBadRequest),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: []any{"record was not processed due to other records failures: " +
					"error REQUIRED_FIELD_MISSING: At least 1 record is required"},
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors:   []any{common.ErrBatchUnprocessedRecord},
					Data:     nil,
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
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errCreatePartial),
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
						"fields": ["LastName"]
					}`)},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					// nolint:lll
					Errors: []any{mockutils.JSONErrorWrapper(`{
						"statusCode": "ALL_OR_NONE_OPERATION_ROLLED_BACK",
						"message": "Record rolled back because not all records were valid and the request was using AllOrNone header",
						"fields": []
					}`)},
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
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
				},
				Then: mockserver.Response(http.StatusBadRequest, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status:       common.BatchStatusFailure,
				Errors:       []any{common.ErrEmptyJSONHTTPResponse},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write with valid payload construction",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
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
					RecordId: "003ak00000luKU1AAM",
					Errors:   nil,
					Data:     nil,
				}, {
					Success:  true,
					RecordId: "003ak00000luKU2AAM",
					Errors:   nil,
					Data:     nil,
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
	updatePayload := testutils.DataFromFile(t, "batch/update/contacts/payload.json")
	errUpdatePartial := testutils.DataFromFile(t, "batch/update/contacts/partial-but-allOrNone.json")
	responseUpdateContacts := testutils.DataFromFile(t, "batch/update/contacts/success.json")

	type record = common.Record

	updateRecords := common.BatchItems{{
		Record: record{
			"id":        "003ak00000jvIfpAAE",
			"FirstName": "Siena",
			"LastName":  "Dyer",
		},
	}, {
		Record: record{
			"id":        "003ak00000jvIfqAAE",
			"FirstName": "Markus",
			"LastName":  "Blevins",
		},
	}}

	tests := []testroutines.BatchWrite{
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errNoIDs),
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
			// For right now no partial response is supported.
			// As of right now, connector always sets payload with AllOrNone=true.
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
					mockcond.BodyBytes(updatePayload),
				},
				Then: mockserver.Response(http.StatusOK, errUpdatePartial),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: nil,
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
						"statusCode": "ALL_OR_NONE_OPERATION_ROLLED_BACK",
						"message": "Record rolled back because not all records were valid and the request was using AllOrNone header",
						"fields": []
					}`)},
					Data: nil,
				}, {
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
						"statusCode": "INVALID_FIELD",
						"message": "No such column 'unknownField_LastName' on sobject of type Contact",
						"fields": []
					}`)},
					Data: nil,
				}},
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write",
			Input: &common.BatchWriteParam{
				ObjectName: "Contact",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/services/data/v60.0/composite/sobjects"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusSuccess,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "003ak00000jvIfpAAE",
					Errors:   nil,
					Data:     nil,
				}, {
					Success:  true,
					RecordId: "003ak00000jvIfqAAE",
					Errors:   nil,
					Data:     nil,
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
