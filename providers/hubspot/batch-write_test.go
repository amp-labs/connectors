package hubspot

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

	errConflictExisting := testutils.DataFromFile(t, "batch/create/contacts/err-conflict.json")
	errManyInvalidFields := testutils.DataFromFile(t, "batch/create/contacts/err-many-invalid-properties.json")
	errPartialSuccess := testutils.DataFromFile(t, "batch/create/contacts/err-partial-success.json")
	responseCreateContacts := testutils.DataFromFile(t, "batch/create/contacts/success.json")

	createRecords := common.BatchItems{{
		Record: map[string]any{
			"email":     "Markus.Blevins@hubspot.com",
			"lastname":  "Blevins",
			"firstname": "Markus",
		},
	}, {
		Record: map[string]any{
			"email":     "Siena.Dyer@hubspot.com",
			"lastname":  "Dyer",
			"firstname": "Siena",
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
				ObjectName: "contacts",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrUnknownWriteType},
		},
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/create"),
				},
				Then: mockserver.Response(http.StatusConflict, errConflictExisting),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status:       common.BatchStatusFailure,
				Errors:       []any{"Contact already exists"},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Bad request without the body",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/create"),
				},
				Then: mockserver.Response(http.StatusBadRequest, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status:       common.BatchStatusFailure,
				Errors:       []any{},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Many errors not traceable to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/create"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errManyInvalidFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: []any{
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"first000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["first000name"]}
					}`),
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"last000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["last000name"]}
					}`),
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"first000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["first000name"]}
					}`),
				},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Partial success where first contact already exists",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/create"),
				},
				Then: mockserver.Response(http.StatusMultiStatus, errPartialSuccess),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusPartial,
				Errors: []any{mockutils.JSONErrorWrapper(`{
					"status": "error",
					"category": "CONFLICT",
					"message": "Contact already exists. Existing ID: 171591000198",
					"context": {"objectWriteTraceId": ["0"], "existingId": ["171591000198"]}
				}`)},
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "",
					Errors: []any{mockutils.JSONErrorWrapper(`{
						"status": "error",
						"category": "CONFLICT",
						"message": "Contact already exists. Existing ID: 171591000198",
						"context": {"objectWriteTraceId": ["0"], "existingId": ["171591000198"]}
					}`)},
					Data: nil,
				}, {
					Success:  true,
					RecordId: "171596044870",
					Errors:   nil,
					Data: map[string]any{
						"email":     "siena.dyer@hubspot.com",
						"firstname": "Siena",
					},
				}},
				SuccessCount: 1,
				FailureCount: 1,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeCreate,
				Batch:      createRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/create"),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusSuccess,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "171596044869",
					Errors:   nil,
					Data: map[string]any{
						"email":     "lilyskinner@hubspot.com",
						"firstname": "Lily",
					},
				}, {
					Success:  true,
					RecordId: "171596044870",
					Errors:   nil,
					Data: map[string]any{
						"email":     "marleyfleming@hubspot.com",
						"firstname": "Marley",
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

	errDuplicateIDs := testutils.DataFromFile(t, "batch/update/contacts/err-duplicate-ids.json")
	errManyInvalidFields := testutils.DataFromFile(t, "batch/update/contacts/err-many-invalid-properties.json")
	errUpdatePartial := testutils.DataFromFile(t, "batch/update/contacts/err-partial-success.json")
	responseUpdateContacts := testutils.DataFromFile(t, "batch/update/contacts/success.json")

	updateRecords := common.BatchItems{{
		Record: map[string]any{
			"id":        "171591000198",
			"email":     "Markus.Blevins@hubspot.com",
			"lastname":  "Blevins (updated)",
			"firstname": "Markus (updated)",
		},
	}, {
		Record: map[string]any{
			"id":        "171591000199",
			"email":     "Siena.Dyer@hubspot.com",
			"firstname": "Siena (updated)",
			"lastname":  "Dyer (updated)",
		},
	}}

	tests := []testroutines.BatchWrite{
		{
			Name: "General high level error not tied to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/update"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errDuplicateIDs),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status:       common.BatchStatusFailure,
				Errors:       []any{"Duplicate IDs found in batch input: [123456]. IDs must be unique"},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Many errors not traceable to any record",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/update"),
				},
				Then: mockserver.Response(http.StatusBadRequest, errManyInvalidFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusFailure,
				Errors: []any{
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"first000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["first000name"]}
					}`),
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"last000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["last000name"]}
					}`),
					mockutils.JSONErrorWrapper(`{
					  "message": "Property \"first000name\" does not exist",
					  "code": "PROPERTY_DOESNT_EXIST",
					  "context": {"propertyName": ["first000name"]}
					}`),
				},
				Results:      nil,
				SuccessCount: 0,
				FailureCount: 2,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Partial result where one contact did not have an id",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeUpdate,
				Batch: common.BatchItems{{
					Record: map[string]any{
						"id":        "unknownIdentifier888", // This identifier will have no response corespondent.
						"email":     "Markus.Blevins@hubspot.com",
						"lastname":  "Blevins (updated)",
						"firstname": "Markus (updated)",
					},
				}, {
					Record: map[string]any{
						"id":        "171591000199",
						"email":     "Siena.Dyer@hubspot.com",
						"firstname": "Siena (updated)",
						"lastname":  "Dyer (updated)",
					},
				}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/update"),
				},
				Then: mockserver.Response(http.StatusMultiStatus, errUpdatePartial),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusPartial,
				Errors: []any{mockutils.JSONErrorWrapper(`{
					  "status": "error",
					  "category": "OBJECT_NOT_FOUND",
					  "message": "Could not get some CONTACT objects, they may be deleted or not exist. Check that ids are valid.",
					  "context": {"ids": [""]}
				}`)},
				Results: []common.WriteResult{{
					Success:  false,
					RecordId: "unknownIdentifier888",
					Errors:   []any{common.ErrBatchUnprocessedRecord},
					Data:     nil,
				}, {
					Success:  true,
					RecordId: "171591000199",
					Errors:   nil,
					Data: map[string]any{
						"firstname": "Siena (updated)",
						"lastname":  "Dyer (updated)",
					},
				}},
				SuccessCount: 1,
				FailureCount: 1,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful write",
			Input: &common.BatchWriteParam{
				ObjectName: "contacts",
				Type:       common.WriteTypeUpdate,
				Batch:      updateRecords,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/batch/update"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetBatchWrite,
			Expected: &common.BatchWriteResult{
				Status: common.BatchStatusSuccess,
				Errors: []any{},
				Results: []common.WriteResult{{
					Success:  true,
					RecordId: "171591000198",
					Errors:   nil,
					Data: map[string]any{
						"firstname": "Markus (updated)",
						"lastname":  "Blevins (updated)",
					},
				}, {
					Success:  true,
					RecordId: "171591000199",
					Errors:   nil,
					Data: map[string]any{
						"firstname": "Siena (updated)",
						"lastname":  "Dyer (updated)",
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
