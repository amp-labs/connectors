package connectwise

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	errorBadRequest := testutils.DataFromFile(t, "write/contacts/bad-request.json")
	errorNotFound := testutils.DataFromFile(t, "write/contacts/not-found.json")
	responseContact := testutils.DataFromFile(t, "write/contacts/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:   "Unknown object",
			Input:  common.WriteParams{ObjectName: "butterfly", RecordData: map[string]any{}},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrResolvingURLPathForObject,
			},
		},
		{
			Name:  "Error invalid payload",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("contact object is invalid: The firstName field is required."),
			},
		},
		{
			Name:  "Error endpoint is not found",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				testutils.StringError("The endpoint does not exist."),
			},
		},
		{
			Name:  "Create task via POST",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "57919",
				Errors:   nil,
				Data: map[string]any{
					"firstName": "Estella",
					"lastName":  "Mcdowell",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact via PUT",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "57919",
				RecordData: map[string]any{
					"lastName": "Sims",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts/57919"),
					mockcond.Body(`{"lastName": "Sims"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "57919",
				Errors:   nil,
				Data: map[string]any{
					"firstName": "Estella",
					"lastName":  "Mcdowell",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact via PATCH due to the payload structure",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "57919",
				RecordData: map[string]any{
					"patch": []any{
						map[string]any{"op": "replace", "path": "/firstName", "value": "Sims"},
						map[string]any{"op": "replace", "path": "/customFields/1/value", "value": true},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts/57919"),
					mockcond.Body(`[
						{"op":"replace","path":"/firstName","value":"Sims"},
						{"op":"replace","path":"/customFields/1/value","value":true}
					]`),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "57919",
				Errors:   nil,
				Data: map[string]any{
					"firstName": "Estella",
					"lastName":  "Mcdowell",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
