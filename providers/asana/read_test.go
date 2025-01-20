package asana

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseAllocations := testutils.DataFromFile(t, "allocations.json")
	responseAllocationsEmpty := testutils.DataFromFile(t, "allocations-empty.json")
	responseGoals := testutils.DataFromFile(t, "goals.json")
	responseMemberships := testutils.DataFromFile(t, "memberships.json")
	responseProjects := testutils.DataFromFile(t, "projects.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "allocations"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},

		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "attributes", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "An Empty response",
			Input: common.ReadParams{ObjectName: "allocations", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAllocationsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},

		{
			Name:  "Read list of all allocations",
			Input: common.ReadParams{ObjectName: "allocations", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAllocations),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"gid":           "12345",
						"resource_type": "allocation",
						"start_date":    "2024-02-28",
						"end_date":      "2024-02-28",
						"effort": map[string]any{
							"type":  "hours",
							"value": float64(54),
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
				}},
				NextPage: "https://app.asana.com/api/1.0/tasks/12345/attachments?limit=2&offset=eyJ0ezI1NiJ9",
				Done:     false,
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Read chosen fileds of goals",
			Input: common.ReadParams{ObjectName: "goals", Fields: connectors.Fields("gid", "resource_type")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseGoals),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"gid":           "12345",
							"resource_type": "goal",
						},
						Raw: map[string]any{
							"gid":           "12345",
							"resource_type": "goal",
							"name":          "Grow web traffic by 30%",
							"owner": map[string]any{
								"gid":           "12345",
								"resource_type": "user",
								"name":          "Greg Sanchez",
							},
						},
					},
				},
				NextPage: "https://app.asana.com/api/1.0/tasks/12345/attachments?limit=2&offset=eyJ0eXAiOUzI1NiJ9",
				Done:     false,
			},
		},
		{
			Name:  "Read list of all memberships",
			Input: common.ReadParams{ObjectName: "memberships", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseMemberships),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":              "12345",
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
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":           "12345",
							"resource_type": "project_membership",
							"parent": map[string]any{
								"gid":           "12345",
								"resource_type": "project",
								"name":          "Stuff to buy",
							},
							"member": map[string]any{
								"gid":           "12345",
								"resource_type": "user",
								"name":          "Greg Sanchez",
							},
							"access_level":     "admin",
							"resource_subtype": "project_membership",
						},
					},
				},
				Done:     false,
				NextPage: "https://app.asana.com/api/1.0/tasks/12345/attachments?limit=2&offset=eQLCJhbGciOiJIUzI1NiJ9",
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Read list of all projects",
			Input: common.ReadParams{ObjectName: "projects", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProjects),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"gid":           "12345",
							"resource_type": "project",
							"name":          "Stuff to buy",
						},
					},
				},
				Done:     false,
				NextPage: "https://app.asana.com/api/1.0/tasks/12345/attachments?limit=2&offset=kjlk",
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
