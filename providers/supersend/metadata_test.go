package supersend

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"unknown"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"unknown": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name: "Successfully describe multiple objects with metadata",
			Input: []string{
				"teams", "senders", "contact/all", "sender-profiles", "org",
				"labels", "campaigns/overview", "managed-domains", "managed-mailboxes",
				"placement-tests", "auto/identitys", "conversation/latest-by-profile",
			},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"teams": {
						DisplayName: "Teams",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"domain": {
								DisplayName:  "domain",
								ValueType:    "string",
								ProviderType: "string",
							},
							"isDefault": {
								DisplayName:  "isDefault",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"senders": {
						DisplayName: "Senders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"warm": {
								DisplayName:  "warm",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"max_per_day": {
								DisplayName:  "max_per_day",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
					"contact/all": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"verified": {
								DisplayName:  "verified",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"sender-profiles": {
						DisplayName: "Sender Profiles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"type": {
								DisplayName:  "type",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"org": {
						DisplayName: "Organization",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"current_plan": {
								DisplayName:  "current_plan",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"labels": {
						DisplayName: "Labels",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"color": {
								DisplayName:  "color",
								ValueType:    "string",
								ProviderType: "string",
							},
							"deleted": {
								DisplayName:  "deleted",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"campaigns/overview": {
						DisplayName: "Campaigns",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"disabled": {
								DisplayName:  "disabled",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"managed-domains": {
						DisplayName: "Managed Domains",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"mailboxCount": {
								DisplayName:  "mailboxCount",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
					"managed-mailboxes": {
						DisplayName: "Managed Mailboxes",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"placement-tests": {
						DisplayName: "Placement Tests",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"auto_send": {
								DisplayName:  "auto_send",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"auto/identitys": {
						DisplayName: "Identities",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"username": {
								DisplayName:  "username",
								ValueType:    "string",
								ProviderType: "string",
							},
							"type": {
								DisplayName:  "type",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
					"conversation/latest-by-profile": {
						DisplayName: "Conversations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
							},
							"is_unread": {
								DisplayName:  "is_unread",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"platform_type": {
								DisplayName:  "platform_type",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
