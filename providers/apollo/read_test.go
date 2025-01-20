package apollo

import (
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

	zeroRecords := testutils.DataFromFile(t, "empty.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	sequencesResponse := testutils.DataFromFile(t, "sequences.json")

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
				Always: mockserver.ResponseString(http.StatusBadRequest, string(unsupportedResponse)),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "opportunity_stages", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, string(zeroRecords)),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Sequences",
			Input: common.ReadParams{
				ObjectName: "sequences",
				Fields:     connectors.Fields("id", "name", "archived"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, string(sequencesResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":       "66e9e215ece19801b219997f",
						"name":     "Target Copywriting Clients in Dublin",
						"archived": false,
					},
					Raw: map[string]any{
						"id":                           "66e9e215ece19801b219997f",
						"name":                         "Target Copywriting Clients in Dublin",
						"archived":                     false,
						"created_at":                   "2024-09-17T20:09:57.837Z",
						"emailer_schedule_id":          "6095a711bd01d100a506d52a",
						"max_emails_per_day":           nil,
						"user_id":                      "66302798d03b9601c7934ebf",
						"same_account_reply_policy_cd": nil,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
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
