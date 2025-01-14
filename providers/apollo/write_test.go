package apollo

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

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	opportunityCreationResponse := testutils.DataFromFile(t, "opportunity-write.json")
	updateDealsResponse := testutils.DataFromFile(t, "update-deals.json")

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
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: "dummy"},
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
			Name: "Successfully creation of an opportunity",
			Input: common.WriteParams{ObjectName: "opportunities", RecordData: map[string]any{
				"name":                 "opportunity - one",
				"amount":               "200",
				"opportunity_stage_id": "65b1974393794c0300d26dcf",
				"closed_date":          "2024-12-18",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, opportunityCreationResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6781406a5dde2401b0dd5ea5",
				Data: map[string]any{
					"id":                  "6781406a5dde2401b0dd5ea5",
					"team_id":             "6508dea16d3b6400a3ed7030",
					"owner_id":            nil,
					"salesforce_owner_id": nil,
					"amount":              float64(200.0),
					"closed_date":         "2024-12-18T00:00:00.000+00:00",
					"account_id":          nil,
					"description":         nil,
					"is_closed":           false,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a deal",
			Input: common.WriteParams{
				ObjectName: "Deals",
				RecordId:   "66d573f1bb530101b230db6f",
				RecordData: map[string]any{
					"amount":               "2500",
					"opportunity_stage_id": "65b1974393794c0300d26dcf",
					"closed_date":          "2024-12-18",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, updateDealsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "66d573f1bb530101b230db6f",
				Data: map[string]any{
					"account_id":              nil,
					"actual_close_date":       nil,
					"amount":                  float64(2500),
					"amount_in_team_currency": float64(2500),
					"closed_date":             "2025-01-10T00:00:00.000+00:00",
					"closed_lost_reason":      nil,
					"closed_won_reason":       nil,
					"created_at":              "2024-09-02T08:14:41.673Z",
					"created_by_id":           "65b17ffc0b8782058df8873f",
					"crm_id":                  nil,
					"crm_owner_id":            nil,
					"crm_record_url":          nil,
					"currency": map[string]any{
						"iso_code": "USD",
						"name":     "US Dollar",
						"symbol":   "$",
					},
					"current_solutions":                nil,
					"deal_probability":                 float64(10),
					"deal_source":                      nil,
					"description":                      nil,
					"exchange_rate_code":               "USD",
					"exchange_rate_value":              float64(1),
					"existence_level":                  "full",
					"forecast_category":                nil,
					"forecasted_revenue":               float64(250),
					"id":                               "66d573f1bb530101b230db6f",
					"is_closed":                        false,
					"is_won":                           false,
					"last_activity_date":               "2025-01-10T16:40:58.745Z",
					"manually_updated_forecast":        nil,
					"manually_updated_probability":     nil,
					"name":                             "Updated Deal Name",
					"next_step":                        nil,
					"next_step_date":                   nil,
					"next_step_last_updated_at":        nil,
					"opportunity_contact_roles":        []any{},
					"opportunity_pipeline_id":          "65b1974393794c0300d26dcd",
					"opportunity_rule_config_statuses": []any{},
					"opportunity_stage_id":             "65b1974393794c0300d26dcf",
					"owner_id":                         nil,
					"probability":                      nil,
					"salesforce_id":                    nil,
					"salesforce_owner_id":              nil,
					"source":                           "api",
					"stage_name":                       nil,
					"stage_updated_at":                 "2024-09-06T12:51:42.143+00:00",
					"team_id":                          "6508dea16d3b6400a3ed7030",
					"typed_custom_fields":              map[string]any{},
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
