package asana

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	respnoseAllocations := testutils.DataFromFile(t, "write-allocations.json")
	responseGoals := testutils.DataFromFile(t, "write-goals.json")
	responseMemberships := testutils.DataFromFile(t, "write-memberships.json")
	responseOrganizationExports := testutils.DataFromFile(t, "write-organizationExports.json")
	responsePortfolios := testutils.DataFromFile(t, "write-portfolios.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "allocations"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "attributes", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Create allocations as POST",
			Input: common.WriteParams{ObjectName: "allocations", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, respnoseAllocations),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"gid":           "12345",
						"resource_type": "allocation",
						"start_date":    "2024-02-28",
						"end_date":      "2024-02-28",

						"effort": map[string]any{
							"type":  "hours",
							"value": float64(50),
						},
						"assignee": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"created_by": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"parent": map[string]any{
							"gid":           "12345",
							"resource_type": "project",
							"name":          "Stuff to buy",
						},
						"resource_subtype": "project_allocation",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create goals as POST",
			Input: common.WriteParams{ObjectName: "goals", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseGoals),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "112233",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"gid":                "112233",
						"resource_type":      "goal",
						"name":               "Grow web traffic by 30%",
						"html_notes":         "<body>Start building brand awareness.</body>",
						"notes":              "Start building brand awareness.",
						"due_on":             "2019-09-15",
						"start_on":           "2019-09-14",
						"is_workspace_level": true,
						"liked":              false,
						"num_likes":          float64(5),
						"team": map[string]any{
							"gid":           "12345",
							"resource_type": "team",
							"name":          "Marketing",
						},
						"workspace": map[string]any{
							"gid":           "12345",
							"resource_type": "workspace",
							"name":          "My Company Workspace",
						},
						"time_period": map[string]any{
							"gid":           "12345",
							"resource_type": "time_period",
							"end_on":        "2019-09-14",
							"start_on":      "2019-09-13",
							"period":        "Q1",
							"display_name":  "Q1 FY22",
						},
						"metric": map[string]any{
							"gid":                   "12345",
							"resource_type":         "task",
							"resource_subtype":      "number",
							"precision":             float64(2),
							"unit":                  "none",
							"currency_code":         "EUR",
							"initial_number_value":  float64(5.2),
							"target_number_value":   float64(10.2),
							"current_number_value":  float64(8.12),
							"current_display_value": "8.12",
							"progress_source":       "manual",
							"is_custom_weight":      false,
							"can_manage":            true,
						},
						"owner": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"current_status_update": map[string]any{
							"gid":              "12345",
							"resource_type":    "status_update",
							"title":            "Status Update - Jun 15",
							"resource_subtype": "project_status_update",
						},
						"status": "green",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create memberships as POST",
			Input: common.WriteParams{ObjectName: "memberships", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseMemberships),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1222",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"gid":              "1222",
						"resource_type":    "membership",
						"resource_subtype": "goal_membership",
						"member": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"parent": map[string]any{
							"gid":           "12345",
							"resource_type": "goal",
							"name":          "Grow web traffic by 30%",
							"owner": map[string]any{
								"gid":           "12345",
								"resource_type": "user",
								"name":          "Greg Sanchez",
							},
						},
						"access_level": "editor",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create organization exports as POST",
			Input: common.WriteParams{ObjectName: "organization_exports", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseOrganizationExports),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "125",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"gid":           "125",
						"resource_type": "organization_export",
						"created_at":    "2012-02-22T02:06:58.147Z",
						"download_url":  "https://asana-export-us-east-1.s3.us",
						"state":         "started",
						"organization": map[string]any{
							"gid":           "12345",
							"resource_type": "workspace",
							"name":          "My Company Workspace",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create portfolios as POST",
			Input: common.WriteParams{ObjectName: "portfolios", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responsePortfolios),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "15",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"gid":           "15",
						"resource_type": "portfolio",
						"name":          "Bug Portfolio",
						"archived":      false,
						"color":         "light-green",
						"created_at":    "2012-02-22T02:06:58.147Z",
						"created_by": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"current_status_update": map[string]any{
							"gid":              "12345",
							"resource_type":    "status_update",
							"title":            "Status Update - Jun 15",
							"resource_subtype": "project_status_update",
						},
						"due_on": "2019-09-15",
						"owner": map[string]any{
							"gid":           "12345",
							"resource_type": "user",
							"name":          "Greg Sanchez",
						},
						"start_on": "2019-09-14",
						"workspace": map[string]any{
							"gid":           "12345",
							"resource_type": "workspace",
							"name":          "My Company Workspace",
						},
						"permalink_url":        "https://app.asana.com/0/resource/123456789/list",
						"public":               false,
						"default_access_level": "viewer",
						"privacy_setting":      "members_only",
						"project_templates": []any{map[string]any{
							"gid":           "12345",
							"resource_type": "project_template",
							"name":          "Packing list",
						}},
					},
				},
			},
			ExpectedErrs: nil,
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
