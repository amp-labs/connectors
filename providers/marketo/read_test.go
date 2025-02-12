package marketo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "zero-records.json")
	unsupportedResponse := testutils.DataFromFile(t, "not-found.json")
	campaignsResponse := testutils.DataFromFile(t, "campaigns.json")
	leadsResponse := testutils.DataFromFile(t, "leads.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "arsenal", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New(string(unsupportedResponse)), //nolint:err113
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "smartcampaigns", Fields: connectors.Fields("description", "id", "name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Campaigns",
			Input: common.ReadParams{
				ObjectName: "campaign",
				Fields:     connectors.Fields("createdAt", "id", "name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, campaignsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"createdat": "2024-08-23T12:09:55Z",
						"id":        float64(1023),
						"name":      "Meme",
					},
					Raw: map[string]any{
						"active":        false,
						"createdAt":     "2024-08-23T12:09:55Z",
						"id":            float64(1023),
						"name":          "Meme",
						"type":          "batch",
						"updatedAt":     "2024-08-23T12:09:55Z",
						"workspaceName": "Default",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		}, {
			Name: "Reading Next Page of Filtered Leads",
			Input: common.ReadParams{
				ObjectName: "leads",
				Fields:     connectors.Fields("email", "id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, leadsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email": "example472@example.com",
						"id":    float64(1),
					},
					Raw: map[string]any{
						"id":        float64(1),
						"firstName": "example472first",
						"lastName":  "example472last",
						"email":     "example472@example.com",
						"updatedAt": "2024-03-20T06:45:20Z",
						"createdAt": "2024-03-20T06:45:12Z",
					},
				}},
				NextPage: "2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
