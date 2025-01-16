package closecrm

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
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	activityResponse := testutils.DataFromFile(t, "activities.json")

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
			Input: common.ReadParams{ObjectName: "united", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				errors.New(string(unsupportedResponse)), //nolint:err113
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "opportunity_stages", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully read activities",
			Input: common.ReadParams{
				ObjectName: "activity",
				Fields:     connectors.Fields("user_id", "user_name", "source", "id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, activityResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":        "acti_4440QLsync5No96XBuQkLfCngfKGXXixrWtWVUjM6lv",
						"source":    "manual",
						"user_id":   "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
						"user_name": "Josep Karage",
					},
					Raw: map[string]any{
						"_type":             "Meeting",
						"activity_at":       "2025-01-20T10:43:46.561000+00:00",
						"actual_duration":   nil,
						"attached_call_ids": []any{},
						"date_created":      "2025-01-20T10:43:46.561000+00:00",
						"date_updated":      "2025-01-14T10:44:08.137000+00:00",
						"ends_at":           "2025-01-20T11:28:46.561000+00:00",
						"id":                "acti_4440QLsync5No96XBuQkLfCngfKGXXixrWtWVUjM6lv",
						"is_recurring":      false,
						"lead_id":           "lead_G7FYn5pkohGlQzAgddQ9zcLOoYSkP2PIFsflsSOMq71",
						"organization_id":   "orga_Bnvb4C6ur74cnEX825rjjDDusTIIpQcFX2qusZcjGr5",
						"source":            "manual",
						"starts_at":         "2025-01-20T10:43:46.561000+00:00",
						"title":             "Contract Finalization",
						"user_id":           "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
						"user_name":         "Josep Karage",
					},
				}},
				Done: true,
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
