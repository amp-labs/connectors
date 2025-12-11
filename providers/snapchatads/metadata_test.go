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
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"email": {
								DisplayName: "email",
								ValueType:   "other",
							},
							"organization_id": {
								DisplayName: "organization_id",
								ValueType:   "other",
							},
							"display_name": {
								DisplayName: "display_name",
								ValueType:   "other",
							},
							"snapchat_username": {
								DisplayName: "snapchat_username",
								ValueType:   "other",
							},
							"member_status": {
								DisplayName: "member_status",
								ValueType:   "other",
							},
						},
					},
					"roles": {
						DisplayName: "Roles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"container_kind": {
								DisplayName: "container_kind",
								ValueType:   "other",
							},
							"container_id": {
								DisplayName: "container_id",
								ValueType:   "other",
							},
							"member_id": {
								DisplayName: "member_id",
								ValueType:   "other",
							},
							"organization_id": {
								DisplayName: "organization_id",
								ValueType:   "other",
							},
							"type": {
								DisplayName: "type",
								ValueType:   "other",
							},
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
