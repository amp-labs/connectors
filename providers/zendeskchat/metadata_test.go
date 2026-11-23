package zendeskchat

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	agentsResponse := testutils.DataFromFile(t, "agents-read.json")
	accountResponse := testutils.DataFromFile(t, "account-read.json")
	chatsResponse := testutils.DataFromFile(t, "chats-read.json")
	emptyListResponse := testutils.DataFromFile(t, "empty-list.json")
	routingSettingsAgentsResponse := testutils.DataFromFile(t, "routing-settings-agents-read.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe list, object and nested list responses",
			Input: []string{"agents", "account", "chats"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v2/chat/agents"),
					Then: mockserver.Response(http.StatusOK, agentsResponse),
				}, {
					If:   mockcond.Path("/api/v2/chat/account"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}, {
					If:   mockcond.Path("/api/v2/chat/chats"),
					Then: mockserver.Response(http.StatusOK, chatsResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"agents": {
						DisplayName: "Agents",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeFloat,
							},
							"email": {
								DisplayName: "email",
								ValueType:   common.ValueTypeString,
							},
							"enabled": {
								DisplayName: "enabled",
								ValueType:   common.ValueTypeBoolean,
							},
							"last_login": {
								DisplayName: "last_login",
								ValueType:   common.ValueTypeOther,
							},
							"roles": {
								DisplayName: "roles",
								ValueType:   common.ValueTypeOther,
							},
						},
						FieldsMap: map[string]string{
							"id":         "id",
							"email":      "email",
							"enabled":    "enabled",
							"last_login": "last_login",
							"roles":      "roles",
						},
					},
					"account": {
						DisplayName: "Account",
						Fields: map[string]common.FieldMetadata{
							"account_key": {
								DisplayName: "account_key",
								ValueType:   common.ValueTypeString,
							},
							"plan": {
								DisplayName: "plan",
								ValueType:   common.ValueTypeOther,
							},
						},
					},
					"chats": {
						DisplayName: "Chats",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"missed": {
								DisplayName: "missed",
								ValueType:   common.ValueTypeBoolean,
							},
							"visitor": {
								DisplayName: "visitor",
								ValueType:   common.ValueTypeOther,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe records keyed by identifier",
			Input: []string{"routing_settings/agents"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v2/chat/routing_settings/agents"),
					Then: mockserver.Response(http.StatusOK, routingSettingsAgentsResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"routing_settings/agents": {
						DisplayName: "Routing_settings/Agents",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"routing": {
								DisplayName: "routing",
								ValueType:   common.ValueTypeOther,
							},
							"skills": {
								DisplayName: "skills",
								ValueType:   common.ValueTypeOther,
							},
						},
						FieldsMap: map[string]string{
							"id":      "id",
							"routing": "routing",
							"skills":  "skills",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Keyed records with empty data cannot be sampled",
			Input: []string{"routing_settings/agents"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v2/chat/routing_settings/agents"),
					Then: mockserver.Response(http.StatusOK, []byte(`{"data": {}}`)),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"routing_settings/agents": common.ErrMissingExpectedValues,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Object with no records cannot be sampled",
			Input: []string{"skills"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v2/chat/skills"),
					Then: mockserver.Response(http.StatusOK, emptyListResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"skills": common.ErrMissingExpectedValues,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test-workspace",
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
