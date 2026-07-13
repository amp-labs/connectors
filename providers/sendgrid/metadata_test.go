package sendgrid

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name: "Successful metadata for email objects",
			Input: []string{
				"contacts",
				"lists",
				"segments",
				"singlesends",
				"templates",
				"field_definitions",
				"verified_senders",
				"senders",
				"bounces",
				"blocks",
				"spam_reports",
				"unsubscribes",
				"invalid_emails",
				"asm_groups",
				"categories",
				"subusers",
				"event_webhook_settings",
				"parse_webhook_settings",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Contact Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"lists": {
						DisplayName: "Lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "List Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"segments": {
						DisplayName: "Segments",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Segment Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"singlesends": {
						DisplayName: "Single Sends",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Single Send Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"templates": {
						DisplayName: "Templates",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Template Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"field_definitions": {
						DisplayName: "Field Definitions",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Field Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"verified_senders": {
						DisplayName: "Verified Senders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Sender Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"nickname": {
								DisplayName:  "Nickname",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"senders": {
						DisplayName: "Senders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Sender Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"nickname": {
								DisplayName:  "Nickname",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"bounces": {
						DisplayName: "Bounces",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"blocks": {
						DisplayName: "Blocks",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"spam_reports": {
						DisplayName: "Spam Reports",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"unsubscribes": {
						DisplayName: "Unsubscribes",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"invalid_emails": {
						DisplayName: "Invalid Emails",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "Created",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"asm_groups": {
						DisplayName: "ASM Groups",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Group Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"categories": {
						DisplayName: "Categories",
						Fields: map[string]common.FieldMetadata{
							"category": {
								DisplayName:  "Category",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"subusers": {
						DisplayName: "Subusers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Subuser Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"username": {
								DisplayName:  "Username",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"event_webhook_settings": {
						DisplayName: "Event Webhook Settings",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Webhook Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"friendly_name": {
								DisplayName:  "Friendly Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"parse_webhook_settings": {
						DisplayName: "Parse Webhook Settings",
						Fields: map[string]common.FieldMetadata{
							"url": {
								DisplayName:  "Url",
								ValueType:    "string",
								ProviderType: "string",
							},
							"hostname": {
								DisplayName:  "Hostname",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:         "Empty objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"lists", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lists": {
						DisplayName: "Lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "List Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"unknown_object": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
			ExpectedErrs: nil,
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
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(serverURL)

	return connector, nil
}
