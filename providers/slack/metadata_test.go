package slack

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	conversationsResponse := testutils.DataFromFile(t, "conversations-list.json")
	usersResponse := testutils.DataFromFile(t, "users-list.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe conversations object with metadata",
			Input: []string{"conversations"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/conversations.list"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, conversationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"conversations": {
						DisplayName: "Conversations",
						Fields: map[string]common.FieldMetadata{
							"id":          {DisplayName: "id", ValueType: common.ValueTypeString},
							"name":        {DisplayName: "name", ValueType: common.ValueTypeString},
							"is_channel":  {DisplayName: "is_channel", ValueType: common.ValueTypeBoolean},
							"is_private":  {DisplayName: "is_private", ValueType: common.ValueTypeBoolean},
							"is_archived": {DisplayName: "is_archived", ValueType: common.ValueTypeBoolean},
							"is_member":   {DisplayName: "is_member", ValueType: common.ValueTypeBoolean},
							"created":     {DisplayName: "created", ValueType: common.ValueTypeFloat},
							"num_members": {DisplayName: "num_members", ValueType: common.ValueTypeFloat},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe users object with metadata",
			Input: []string{"users"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/users.list"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id":        {DisplayName: "id", ValueType: common.ValueTypeString},
							"team_id":   {DisplayName: "team_id", ValueType: common.ValueTypeString},
							"name":      {DisplayName: "name", ValueType: common.ValueTypeString},
							"deleted":   {DisplayName: "deleted", ValueType: common.ValueTypeBoolean},
							"real_name": {DisplayName: "real_name", ValueType: common.ValueTypeString},
							"is_admin":  {DisplayName: "is_admin", ValueType: common.ValueTypeBoolean},
							"is_bot":    {DisplayName: "is_bot", ValueType: common.ValueTypeBoolean},
							"tz_offset": {DisplayName: "tz_offset", ValueType: common.ValueTypeFloat},
							"updated":   {DisplayName: "updated", ValueType: common.ValueTypeFloat},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe multiple objects",
			Input: []string{"conversations", "users"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/api/conversations.list"),
						Then: mockserver.Response(http.StatusOK, conversationsResponse),
					},
					{
						If:   mockcond.Path("/api/users.list"),
						Then: mockserver.Response(http.StatusOK, usersResponse),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"conversations": {
						DisplayName: "Conversations",
						Fields: map[string]common.FieldMetadata{
							"id":   {DisplayName: "id", ValueType: common.ValueTypeString},
							"name": {DisplayName: "name", ValueType: common.ValueTypeString},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id":   {DisplayName: "id", ValueType: common.ValueTypeString},
							"name": {DisplayName: "name", ValueType: common.ValueTypeString},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// Preserve the /api path from the Slack base URL when redirecting to the mock server.
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
