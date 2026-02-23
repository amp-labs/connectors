package expensify

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

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	responsePolicy := testutils.DataFromFile(t, "write-policy.json")
	responseReport := testutils.DataFromFile(t, "write-report.json")
	responseAuthError := testutils.DataFromFile(t, "error-auth.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "policy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Unsupported object returns operation not supported error",
			Input: common.WriteParams{
				ObjectName: "invoices",
				RecordData: map[string]any{"type": "invoice"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "API authentication error is propagated",
			Input: common.WriteParams{
				ObjectName: "policy",
				RecordData: map[string]any{"type": "policy", "policyName": "Test"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAuthError),
			}.Server(),
			ExpectedErrs: []error{common.ErrRequestFailed},
		},
		{
			Name: "Successfully create a policy",
			Input: common.WriteParams{
				ObjectName: "policy",
				RecordData: map[string]any{
					"type":       "policy",
					"policyName": "My New Policy",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responsePolicy),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1A071BFDDFA06534",
				Errors:   nil,
				Data: map[string]any{
					"policyID":     "1A071BFDDFA06534",
					"policyName":   "My New Policy",
					"responseCode": float64(200),
				},
			},
		},
		{
			Name: "Successfully create a report",
			Input: common.WriteParams{
				ObjectName: "report",
				RecordData: map[string]any{
					"type":     "report",
					"policyID": "94C31405ED1893F1",
					"report": map[string]any{
						"title": "Name of the report",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseReport),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "R005qyoxJjMV",
				Errors:   nil,
				Data: map[string]any{
					"reportID":     "R005qyoxJjMV",
					"reportName":   "Name of the report",
					"responseCode": float64(200),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
