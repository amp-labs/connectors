package phoneburner

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	got, err := conn.ListObjectMetadata(
		t.Context(),
		[]string{"contacts", "tags", "customfields", "members"},
	)
	if err != nil {
		t.Fatalf("ListObjectMetadata returned error: %v", err)
	}

	want := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"contacts": {
				DisplayName: "Contacts",
				Fields: map[string]common.FieldMetadata{
					"id": {
						DisplayName:  "Id",
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
			"tags": {
				DisplayName: "Tags",
				Fields: map[string]common.FieldMetadata{
					"id": {
						DisplayName:  "Id",
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
			"customfields": {
				DisplayName: "Custom Fields",
				Fields: map[string]common.FieldMetadata{
					"id": {
						DisplayName:  "Id",
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
			"members": {
				DisplayName: "Members",
				Fields: map[string]common.FieldMetadata{
					"id": {
						DisplayName:  "Id",
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
		},
		Errors: nil,
	}

	if !testroutines.ComparatorSubsetMetadata("", got, want) {
		t.Fatalf("metadata result mismatch: expected subset not found")
	}
}

