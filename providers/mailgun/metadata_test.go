package mailgun

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
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
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"unknown": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name: "Successfully describe multiple objects with metadata",
			Input: []string{
				"domains",
				"templates",
				"lists",
				"webhooks",
				"users",
				"domains/dynamic_pools/assignable",
				"ip_warmups",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"domains": {
						DisplayName: "Domains",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"is_disabled": {
								DisplayName:  "is_disabled",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
						},
					},
					"templates": {
						DisplayName: "Templates",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"domain": {
								DisplayName:  "domain",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "description",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"lists": {
						DisplayName: "Lists",
						Fields: map[string]common.FieldMetadata{
							"address": {
								DisplayName:  "address",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"members_count": {
								DisplayName:  "members_count",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
							"access_level": {
								DisplayName:  "access_level",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"webhooks": {
						DisplayName: "Webhooks",
						Fields: map[string]common.FieldMetadata{
							"webhook_id": {
								DisplayName:  "webhook_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"url": {
								DisplayName:  "url",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "description",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"event_types": {
								DisplayName:  "event_types",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"account_id": {
								DisplayName:  "account_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"activated": {
								DisplayName:  "activated",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
						},
					},
					"domains/dynamic_pools/assignable": {
						DisplayName: "Assignable Domains",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"account_id": {
								DisplayName:  "account_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"ip_warmups": {
						DisplayName: "IP Address Warmup",
						Fields: map[string]common.FieldMetadata{
							"ip": {
								DisplayName:  "ip",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"stage_number": {
								DisplayName:  "stage_number",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
							"stage_start_volume": {
								DisplayName:  "stage_start_volume",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
							"stage_volume_limit": {
								DisplayName:  "stage_volume_limit",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
