package snapchatads

import (
	"context"
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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	organizationIdResponse := testutils.DataFromFile(t, "organization-id.json")
	membersResponse := testutils.DataFromFile(t, "members.json")
	rolesResponse := testutils.DataFromFile(t, "roles.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"members", "roles"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/members"),
					Then: mockserver.Response(http.StatusOK, membersResponse),
				}, {
					If:   mockcond.Path("/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/roles"),
					Then: mockserver.Response(http.StatusOK, rolesResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"members": {
						DisplayName: "Members",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":                "id",
							"email":             "email",
							"organization_id":   "organization_id",
							"display_name":      "display_name",
							"snapchat_username": "snapchat_username",
							"member_status":     "member_status",
						},
					},
					"roles": {
						DisplayName: "Roles",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":              "id",
							"container_kind":  "container_kind",
							"container_id":    "container_id",
							"member_id":       "member_id",
							"organization_id": "organization_id",
							"type":            "type",
						},
					},
				},
				Errors: nil,
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

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	_, err = connector.GetPostAuthInfo(context.Background())
	if err != nil {
		return nil, err
	}

	return connector, nil
}
