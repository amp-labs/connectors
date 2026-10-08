package lob

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	writeAddressesResponse := testutils.DataFromFile(t, "write/addresses.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "addresses"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Create address",
			Input: common.WriteParams{ObjectName: "addresses", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/addresses"),
				},
				Then: mockserver.Response(http.StatusOK, writeAddressesResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "adr_62882589e176345e",
				Data: map[string]any{
					"name":  "CAMYLLE KOELPIN",
					"email": "arvidschaden@collins.org",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for object, method := range map[string]string{
		"billing_groups":              http.MethodPost,
		"buckslips":                   http.MethodPatch,
		"campaigns":                   http.MethodPatch,
		"cards":                       http.MethodPost,
		"informed_delivery_campaigns": http.MethodPatch,
		"links":                       http.MethodPatch,
		"templates":                   http.MethodPost,
		"uploads":                     http.MethodPatch,
	} {
		tests = append(tests, testconn.TestCaseWrite{
			Name:  "Update object: " + object,
			Input: common.WriteParams{ObjectName: object, RecordId: "7", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(method),
					mockcond.Path("/v1/" + object + "/7"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"id": "12345"}`),
			}.Server(),
			Comparator:   testconn.ComparatorSubsetWrite,
			Expected:     &common.WriteResult{Success: true, RecordId: "12345", Data: map[string]any{"id": "12345"}},
			ExpectedErrs: nil,
		})
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
