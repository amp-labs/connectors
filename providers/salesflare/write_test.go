package salesflare

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

	errorBadRequest := testutils.DataFromFile(t, "write/bad-request.json")
	responseContact := testutils.DataFromFile(t, "write/contacts-new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Bad request from provider",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError(
					`"value" must contain at least one of [name, email, prefix, firstname, middle, lastname, suffix]`,
				),
			},
		},
		{
			Name:  "Create contact",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/contacts"),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "285528841",
				Errors:   nil,
				Data:     map[string]any{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update contact",
			Input: common.WriteParams{ObjectName: "contacts", RecordId: "285528841", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/contacts/285528841"),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "285528841",
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
