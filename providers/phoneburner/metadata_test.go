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
		[]string{"contacts", "folders", "members", "voicemails", "phonenumber", "dialsession", "customfields"},
	)
	if err != nil {
		t.Fatalf("ListObjectMetadata returned error: %v", err)
	}

	want := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"contacts": {
				DisplayName: "Contacts",
				Fields: map[string]common.FieldMetadata{
					"contact_user_id": {
						DisplayName:  "Contact User Id",
						ValueType:    "string",
						ProviderType: "string",
					},
					"first_name": {
						DisplayName:  "First Name",
						ValueType:    "string",
						ProviderType: "string",
					},
				},
			},
			"folders": {
				DisplayName: "Folders",
				Fields: map[string]common.FieldMetadata{
					"folder_id": {
						DisplayName:  "Folder Id",
						ValueType:    "string",
						ProviderType: "string",
					},
					"folder_name": {
						DisplayName:  "Folder Name",
						ValueType:    "string",
						ProviderType: "string",
					},
				},
			},
			"members": {
				DisplayName: "Members",
				Fields: map[string]common.FieldMetadata{
					"user_id": {
						DisplayName:  "User Id",
						ValueType:    "string",
						ProviderType: "string",
					},
					"email_address": {
						DisplayName:  "Email Address",
						ValueType:    "string",
						ProviderType: "string",
					},
				},
			},
			"voicemails": {
				DisplayName: "Voicemails",
				Fields: map[string]common.FieldMetadata{
					"recording_id": {
						DisplayName:  "Recording Id",
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
			"phonenumber": {
				DisplayName: "Phone Number",
				Fields: map[string]common.FieldMetadata{
					"phone_number": {
						DisplayName:  "Phone Number",
						ValueType:    "string",
						ProviderType: "string",
					},
				},
			},
			"dialsession": {
				DisplayName: "Dial Sessions",
				Fields: map[string]common.FieldMetadata{
					"dialsession_id": {
						DisplayName:  "Dialsession Id",
						ValueType:    "string",
						ProviderType: "string",
					},
					"start_when": {
						DisplayName:  "Start When",
						ValueType:    "datetime",
						ProviderType: "datetime",
					},
				},
			},
			"customfields": {
				DisplayName: "Custom Fields",
				Fields: map[string]common.FieldMetadata{
					"custom_field_id": {
						DisplayName:  "Custom Field Id",
						ValueType:    "string",
						ProviderType: "string",
					},
					"display_name": {
						DisplayName:  "Display Name",
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

func TestListObjectMetadata_EmptyObjects(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	_, err = conn.ListObjectMetadata(t.Context(), nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != common.ErrMissingObjects {
		t.Fatalf("expected %v, got %v", common.ErrMissingObjects, err)
	}
}

func TestListObjectMetadata_UnsupportedObject(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	got, err := conn.ListObjectMetadata(t.Context(), []string{"contacts", "unknown_object"})
	if err != nil {
		t.Fatalf("ListObjectMetadata returned error: %v", err)
	}

	want := &common.ListObjectMetadataResult{
		Result: map[string]common.ObjectMetadata{
			"contacts": {
				DisplayName: "Contacts",
				Fields: map[string]common.FieldMetadata{
					"contact_user_id": {
						DisplayName:  "Contact User Id",
						ValueType:    "string",
						ProviderType: "string",
					},
				},
			},
		},
		Errors: map[string]error{
			"unknown_object": common.ErrObjectNotSupported,
		},
	}

	if !testroutines.ComparatorSubsetMetadata("", got, want) {
		t.Fatalf("metadata result mismatch: expected subset not found")
	}
}
