package closecrm

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

	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	leadCreationResponse := testutils.DataFromFile(t, "create-lead.json")
	updateLeadResponse := testutils.DataFromFile(t, "update-lead.json")

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
				errors.New(string(unsupportedResponse)),
			},
		},
		{
			Name: "Successfully creation of a lead",
			Input: common.WriteParams{ObjectName: "lead", RecordData: map[string]any{
				"name":        "Bluth Company",
				"url":         "http://thebluthcompany.tumblr.com/",
				"description": "Best. Show. Ever.",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, leadCreationResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "lead_q6XxGP7gFqn1UZ2F8lrmm0JuIOQzEsUGoVY0Yz9fLWx",
				Data: map[string]any{
					"addresses":       []any{},
					"contacts":        []any{},
					"created_by":      "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
					"created_by_name": "Josep Karage",
					"custom":          map[string]any{},
					"date_created":    "2025-01-16T07:48:13.837000+00:00",
					"date_updated":    "2025-01-16T07:48:13.837000+00:00",
					"description":     "Best. Show. Ever.",
					"display_name":    "Bluth Company",
					"html_url":        "https://app.close.com/lead/lead_q6XxGP7gFqn1UZ2F8lrmm0JuIOQzEsUGoVY0Yz9fLWx/",
					"id":              "lead_q6XxGP7gFqn1UZ2F8lrmm0JuIOQzEsUGoVY0Yz9fLWx",
					"name":            "Bluth Company",
					"opportunities":   []any{},
					"organization_id": "orga_Bnvb4C6ur74cnEX825rjjDDusTIIpQcFX2qusZcjGr5",
					"status_id":       "stat_s4Pu3Yd7fVHaEqNLyVaDAbf72PRbhkI9UDgfgRvmKw5",
					"status_label":    "Potential",
					"tasks":           []any{},
					"updated_by":      "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
					"updated_by_name": "Josep Karage",
					"url":             "http://thebluthcompany.tumblr.com/",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a lead",
			Input: common.WriteParams{
				ObjectName: "lead",
				RecordId:   "lead_hVUTYHtMmG2p7DNRJGy3IQuB3GfBLwCu46qsw1gbm6c",
				RecordData: map[string]any{
					"url": "http://thebluthcompany.pumblr.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, updateLeadResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "lead_hVUTYHtMmG2p7DNRJGy3IQuB3GfBLwCu46qsw1gbm6c",
				Data: map[string]any{
					"addresses":       []any{},
					"contacts":        []any{},
					"created_by":      "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
					"created_by_name": "Josep Karage",
					"custom":          map[string]any{},
					"date_created":    "2025-01-16T08:00:22.455000+00:00",
					"date_updated":    "2025-01-16T08:01:37.941000+00:00",
					"description":     "Best. Show. Ever.",
					"display_name":    "Bluth Company",
					"html_url":        "https://app.close.com/lead/lead_hVUTYHtMmG2p7DNRJGy3IQuB3GfBLwCu46qsw1gbm6c/",
					"id":              "lead_hVUTYHtMmG2p7DNRJGy3IQuB3GfBLwCu46qsw1gbm6c",
					"name":            "Bluth Company",
					"opportunities":   []any{},
					"organization_id": "orga_Bnvb4C6ur74cnEX825rjjDDusTIIpQcFX2qusZcjGr5",
					"status_id":       "stat_s4Pu3Yd7fVHaEqNLyVaDAbf72PRbhkI9UDgfgRvmKw5",
					"status_label":    "Potential",
					"tasks":           []any{},
					"updated_by":      "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
					"updated_by_name": "Josep Karage",
					"url":             "http://thebluthcompany.pumblr.com",
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
