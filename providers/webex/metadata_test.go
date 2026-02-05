package webex

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	peopleResponse := testutils.DataFromFile(t, "read-people.json")
	rolesResponse := testutils.DataFromFile(t, "read-roles.json")
	groupsResponse := testutils.DataFromFile(t, "read-groups.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe people object by sampling first record",
			Input: []string{"people"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/people"),
					mockcond.QueryParam("max", "1"),
				},
				Then: mockserver.Response(http.StatusOK, peopleResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"people": {
						DisplayName: "People",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"emails": {
								DisplayName:  "emails",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"displayName": {
								DisplayName:  "displayName",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"nickName": {
								DisplayName:  "nickName",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"lastName": {
								DisplayName:  "lastName",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"orgId": {
								DisplayName:  "orgId",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"type": {
								DisplayName:  "type",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe roles object by sampling first record",
			Input: []string{"roles"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/roles"),
					mockcond.QueryParam("max", "1"),
				},
				Then: mockserver.Response(http.StatusOK, rolesResponse),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"roles": {
						DisplayName: "Roles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe groups object by sampling first record",
			Input: []string{"groups"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/groups"),
					mockcond.QueryParam("count", "1"),
				},
				Then: mockserver.Response(http.StatusOK, groupsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"groups": {
						DisplayName: "Groups",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"displayName": {
								DisplayName:  "displayName",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"orgId": {
								DisplayName:  "orgId",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"created": {
								DisplayName:  "created",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"lastModified": {
								DisplayName:  "lastModified",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when server responds with 500",
			Input: []string{"people"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/people"),
					mockcond.QueryParam("max", "1"),
				},
				Then: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"people": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
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

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
