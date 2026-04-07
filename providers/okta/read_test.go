package okta

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseUsers := testutils.DataFromFile(t, "read-users.json")
	responseGroups := testutils.DataFromFile(t, "read-groups.json")
	responseApps := testutils.DataFromFile(t, "read-apps.json")
	responseUsersWithCustomFields := testutils.DataFromFile(t, "read-users-with-custom-fields.json")
	responseGroupsWithCustomFields := testutils.DataFromFile(t, "read-groups-with-custom-fields.json")
	errorResponse := testutils.DataFromFile(t, "error.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read users successfully",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "status", "profile"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
						Raw: map[string]any{
							"id":      "00u1234567890abcdef",
							"status":  "ACTIVE",
							"created": "2023-01-15T10:00:00.000Z",
							"profile": map[string]any{
								"firstName":   "John",
								"lastName":    "Doe",
								"email":       "john.doe@example.com",
								"login":       "john.doe@example.com",
								"mobilePhone": "+14155551234",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
						Raw: map[string]any{
							"id":      "00u9876543210zyxwvu",
							"status":  "PROVISIONED",
							"created": "2023-02-20T11:00:00.000Z",
							"profile": map[string]any{
								"firstName": "Jane",
								"lastName":  "Smith",
								"email":     "jane.smith@example.com",
								"login":     "jane.smith@example.com",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with Link header pagination",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "status"),
			},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link", `<https://trial-4378019.okta.com/api/v1/users?limit=200&after=xyz123>; rel="next"`),
					mockserver.Response(http.StatusOK, responseUsers),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
						Raw: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
					},
					{
						Fields: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
						Raw: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
					},
				},
				NextPage: "https://trial-4378019.okta.com/api/v1/users?limit=200&after=xyz123",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read groups successfully",
			Input: common.ReadParams{
				ObjectName: "groups",
				Fields:     connectors.Fields("id", "type", "profile"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseGroups),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":   "00g1234567890abcdef",
							"type": "OKTA_GROUP",
						},
						Raw: map[string]any{
							"id":      "00g1234567890abcdef",
							"type":    "OKTA_GROUP",
							"created": "2023-01-10T09:00:00.000Z",
							"profile": map[string]any{
								"name":        "Engineering",
								"description": "Engineering team members",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read apps successfully",
			Input: common.ReadParams{
				ObjectName: "apps",
				Fields:     connectors.Fields("id", "name", "label", "status"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseApps),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "0oa1234567890abcdef",
							"name":   "okta_org2org",
							"label":  "Okta Org2Org",
							"status": "ACTIVE",
						},
						Raw: map[string]any{
							"id":          "0oa1234567890abcdef",
							"name":        "okta_org2org",
							"label":       "Okta Org2Org",
							"status":      "ACTIVE",
							"signOnMode":  "SAML_2_0",
							"created":     "2023-01-10T09:00:00.000Z",
							"lastUpdated": "2023-12-15T10:00:00.000Z",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with incremental sync (Since filter)",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "status"),
				Since:      time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
						Raw: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
					},
					{
						Fields: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
						Raw: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read returns error on bad request",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name: "Read users with custom fields flattened from profile",
			Input: common.ReadParams{
				ObjectName: "users",
				// Custom field test_custom_field is flattened from profile to root level
				Fields: connectors.Fields("id", "email", "firstName", "test_custom_field"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsersWithCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":                "00u1234567890abcdef",
							"email":             "john.doe@example.com",
							"firstname":         "John",
							"test_custom_field": "custom_value_1",
						},
						Raw: map[string]any{
							"id":     "00u1234567890abcdef",
							"status": "ACTIVE",
						},
					},
					{
						Fields: map[string]any{
							"id":                "00u9876543210zyxwvu",
							"email":             "jane.smith@example.com",
							"firstname":         "Jane",
							"test_custom_field": "custom_value_2",
						},
						Raw: map[string]any{
							"id":     "00u9876543210zyxwvu",
							"status": "PROVISIONED",
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read groups with custom fields flattened from profile",
			Input: common.ReadParams{
				ObjectName: "groups",
				// Custom field test_custom_field_group is flattened from profile to root level
				Fields: connectors.Fields("id", "name", "description", "test_custom_field_group"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseGroupsWithCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":                      "00g1234567890abcdef",
							"name":                    "Engineering",
							"description":             "Engineering team members",
							"test_custom_field_group": "engineering_custom",
						},
						Raw: map[string]any{
							"id":   "00g1234567890abcdef",
							"type": "OKTA_GROUP",
						},
					},
					{
						Fields: map[string]any{
							"id":                      "00g9876543210zyxwvu",
							"name":                    "Marketing",
							"description":             "Marketing team members",
							"test_custom_field_group": "marketing_custom",
						},
						Raw: map[string]any{
							"id":   "00g9876543210zyxwvu",
							"type": "OKTA_GROUP",
						},
					},
				},
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
