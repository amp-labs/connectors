package marketo

import (
	"errors"
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

	unsupportedResponse := testutils.DataFromFile(t, "not-found.json")
	leadCreationResponse := testutils.DataFromFile(t, "write-leads.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name:  "Unsupported object",
			Input: common.WriteParams{ObjectName: "lalala", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				errors.New(string(unsupportedResponse)), // nolint:goerr113
			},
		},
		{
			Name: "Successfully creation of a lead",
			Input: common.WriteParams{ObjectName: "leads", RecordData: map[string]any{
				"input": []map[string]any{
					{
						"email":     "joseph@gmail.com",
						"firstName": "Example Lead",
					},
				},
				"action":      "createOnly",
				"lookupField": "email",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/v1/leads.json"),
				},
				Then: mockserver.Response(http.StatusOK, leadCreationResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "550",
				Data: map[string]any{
					"id":     float64(550),
					"status": "created",
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
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
