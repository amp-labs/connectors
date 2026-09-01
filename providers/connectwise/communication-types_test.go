package connectwise

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// TestAttachCommunicationItems verifies that Read of a contact
// correctly expose virtual fields (Email, Phone, Fax) from communication items.
func TestAttachCommunicationItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "Empty input",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "No communication items",
			input: map[string]any{
				"firstName": "John",
			},
			expected: map[string]any{
				"firstName": "John",
			},
		},
		{
			name: "Single default email",
			input: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
				},
			},
			expected: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
				},
				"AMPERSAND-defaultEmail":   "john@example.com",
				"AMPERSAND-defaultEmailId": "1",
			},
		},
		{
			name: "Multiple communication items, only one default",
			input: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
					map[string]any{
						"communicationType": "Email",
						"value":             "john2@example.com",
						"defaultFlag":       false,
						"type": map[string]any{
							"id": 2,
						},
					},
				},
			},
			expected: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
					map[string]any{
						"communicationType": "Email",
						"value":             "john2@example.com",
						"defaultFlag":       false,
						"type": map[string]any{
							"id": 2,
						},
					},
				},
				"AMPERSAND-defaultEmail":   "john@example.com",
				"AMPERSAND-defaultEmailId": "1",
			},
		},
		{
			name: "All types of communication items",
			input: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
					map[string]any{
						"communicationType": "Phone",
						"value":             "123456789",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 2,
						},
					},
					map[string]any{
						"communicationType": "Fax",
						"value":             "987654321",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 3,
						},
					},
				},
			},
			expected: map[string]any{
				"communicationItems": []any{
					map[string]any{
						"communicationType": "Email",
						"value":             "john@example.com",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 1,
						},
					},
					map[string]any{
						"communicationType": "Phone",
						"value":             "123456789",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 2,
						},
					},
					map[string]any{
						"communicationType": "Fax",
						"value":             "987654321",
						"defaultFlag":       true,
						"type": map[string]any{
							"id": 3,
						},
					},
				},
				"AMPERSAND-defaultEmail":   "john@example.com",
				"AMPERSAND-defaultEmailId": "1",
				"AMPERSAND-defaultPhone":   "123456789",
				"AMPERSAND-defaultPhoneId": "2",
				"AMPERSAND-defaultFax":     "987654321",
				"AMPERSAND-defaultFaxId":   "3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := jsonquery.Convertor.NodeFromMap(tt.input)
			if err != nil {
				t.Fatalf("failed to convert map to node: %v", err)
			}

			err = attachCommunicationItems(node, tt.input)
			if err != nil {
				t.Fatalf("attachCommunicationItems failed: %v", err)
			}

			res := testutils.NewCompareResult()
			res.Assert("record", tt.expected, tt.input)
			res.Validate(t, tt.name)
		})
	}
}

// TestPostPayloadWithCommunicationItems verifies that creating a contact with virtual communication fields
// correctly transforms them into the expected communicationItems structure.
func TestPostPayloadWithCommunicationItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		record      common.Record
		server      *httptest.Server
		expected    common.Record
		expectedErr error
	}{
		{
			name:     "Nil record",
			record:   nil,
			server:   mockserver.Dummy(),
			expected: nil,
		},
		{
			name:     "Empty record",
			record:   common.Record{},
			server:   mockserver.Dummy(),
			expected: common.Record{},
		},
		{
			name: "No communication virtual fields",
			record: common.Record{
				"firstName": "John",
			},
			server: mockserver.Dummy(),
			expected: common.Record{
				"firstName": "John",
			},
		},
		{
			name: "All communication virtual fields and IDs provided",
			record: common.Record{
				"firstName":                "John",
				"AMPERSAND-defaultEmail":   "john@example.com",
				"AMPERSAND-defaultEmailId": "1",
				"AMPERSAND-defaultPhone":   "123456789",
				"AMPERSAND-defaultPhoneId": "2",
				"AMPERSAND-defaultFax":     "987654321",
				"AMPERSAND-defaultFaxId":   "3",
			},
			server: mockserver.Dummy(),
			expected: common.Record{
				"firstName": "John",
				"communicationItems": []createCommunicationItemPayload{
					{
						Type:              communicationItemTypePayload{Id: "1"},
						Value:             "john@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
					{
						Type:              communicationItemTypePayload{Id: "2"},
						Value:             "123456789",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
					{
						Type:              communicationItemTypePayload{Id: "3"},
						Value:             "987654321",
						DefaultFlag:       true,
						CommunicationType: "Fax",
					},
				},
			},
		},
		{
			name: "Communication virtual fields provided, missing IDs (fetching needed)",
			record: common.Record{
				"AMPERSAND-defaultEmail": "john@example.com",
				"AMPERSAND-defaultPhone": "123456789",
			},
			server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
					mockcond.QueryParam("conditions", "defaultFlag=true"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `[
					{"id": 10, "emailFlag": true, "defaultFlag": true},
					{"id": 20, "phoneFlag": true, "defaultFlag": true}
				]`),
			}.Server(),
			expected: common.Record{
				"communicationItems": []createCommunicationItemPayload{
					{
						Type:              communicationItemTypePayload{Id: "10"},
						Value:             "john@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
					{
						Type:              communicationItemTypePayload{Id: "20"},
						Value:             "123456789",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
				},
			},
		},
		{
			name: "Mixed case: some IDs provided, some missing",
			record: common.Record{
				"AMPERSAND-defaultEmail":   "john@example.com",
				"AMPERSAND-defaultEmailId": "1",
				"AMPERSAND-defaultPhone":   "123456789",
			},
			server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
					mockcond.QueryParam("conditions", "defaultFlag=true"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `[
					{"id": 20, "phoneFlag": true, "defaultFlag": true}
				]`),
			}.Server(),
			expected: common.Record{
				"communicationItems": []createCommunicationItemPayload{
					{
						Type:              communicationItemTypePayload{Id: "1"},
						Value:             "john@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
					{
						Type:              communicationItemTypePayload{Id: "20"},
						Value:             "123456789",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
				},
			},
		},
		{
			name: "Invalid virtual field type",
			record: common.Record{
				"AMPERSAND-defaultEmail": 123,
			},
			server:      mockserver.Dummy(),
			expected:    common.Record{"AMPERSAND-defaultEmail": 123},
			expectedErr: common.ErrInvalidVirtualField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connector, err := constructTestConnector(tt.server)
			if err != nil {
				t.Fatalf("failed to construct connector: %v", err)
			}

			err = connector.postPayloadWithCommunicationItems(t.Context(), tt.record)

			res := testutils.NewCompareResult()
			res.AssertErr("error", tt.expectedErr, err)
			res.Assert("record", tt.expected, tt.record)
			res.Validate(t, tt.name)
		})
	}
}

// TestContactsFullUpdatePayload verifies that a full contact update (PUT-like) correctly wipes
// omitted fields and synchronizes communication items by adding, updating, or removing them.
func TestContactsFullUpdatePayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		payload                common.Record
		respContact            string
		respCommunicationItems string
		expected               []patchOperationPayload
	}{
		{
			name:    "Nil record",
			payload: nil,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
			},
		},
		{
			name:    "Empty record",
			payload: common.Record{},
			respContact: `{
				"id": 123,
				"firstName": "John",
				"lastName": "Doe"
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "remove", Path: "lastName"},
			},
		},
		{
			name: "Core fields update, wiping others",
			payload: common.Record{
				"firstName": "Jane",
				"title":     "Manager",
			},
			respContact: `{
				"id": 123,
				"firstName": "John",
				"lastName": "Doe",
				"department": "Sales"
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "remove", Path: "department"},
				{Op: "remove", Path: "lastName"},
				{Op: "replace", Path: "firstName", Value: "Jane"},
				{Op: "replace", Path: "title", Value: "Manager"},
			},
		},
		{
			name: "Custom fields replacement",
			payload: common.Record{
				"firstName":     "Fred",
				"customField83": "Skiing",
			},
			respContact: `{
				"id": 123,
				"firstName": "John",
				"customFields": [
					{"id": 53, "value": "Old value"}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "replace", Path: "customField83", Value: "Skiing"},
				{Op: "replace", Path: "firstName", Value: "Fred"},
			},
		},
		{
			name: "Communication items: add, replace, remove",
			payload: common.Record{
				"firstName":                "Johny",
				"AMPERSAND-defaultEmail":   "new@example.com",
				"AMPERSAND-defaultEmailId": "100", // Existing type ID
				"AMPERSAND-defaultPhone":   "555-1234",
				"AMPERSAND-defaultPhoneId": "300", // New type ID
			},
			respContact: `{
				"id": 123,
				"firstName": "John",
				"communicationItems": [
					{
						"communicationType": "Email",
						"value": "old@example.com",
						"defaultFlag": true,
						"type": {"id": 100}
					},
					{
						"communicationType": "Fax",
						"value": "111-2222",
						"defaultFlag": false,
						"type": {"id": 200}
					}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "replace", Path: "firstName", Value: "Johny"},
				{Op: "replace", Path: "/communicationItems/0/value", Value: "new@example.com"},
				{
					Op:   "add",
					Path: "/communicationItems/2",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "300"},
						Value:             "555-1234",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
				},
				{Op: "remove", Path: "/communicationItems/1"}, // Fax is not in payload, hence to be removed.
			},
		},
		{
			name: "Fetch missing communication IDs",
			payload: common.Record{
				"firstName":              "Johny",
				"AMPERSAND-defaultEmail": "fetch@example.com",
			},
			respContact: `{
				"id": 123,
				"firstName": "John"
			}`,
			respCommunicationItems: `[
				{"id": 5, "emailFlag": true, "defaultFlag": true}
			]`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "replace", Path: "firstName", Value: "Johny"},
				{
					Op:   "add",
					Path: "/communicationItems/0",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "5"},
						Value:             "fetch@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
				},
			},
		},
		{
			name: "Complex full update: many removals and additions",
			payload: common.Record{
				"firstName":                "Multiple",
				"AMPERSAND-defaultEmail":   "new.default@test.com",
				"AMPERSAND-defaultEmailId": "100",
				"AMPERSAND-defaultPhone":   "555-9999",
				"AMPERSAND-defaultPhoneId": "200",
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{"id": 10, "communicationType": "Email", "value": "e1@test.com", "defaultFlag": true, "type": {"id": 1}},
					{"id": 11, "communicationType": "Email", "value": "e2@test.com", "defaultFlag": false, "type": {"id": 2}},
					{"id": 20, "communicationType": "Phone", "value": "555-1", "defaultFlag": true, "type": {"id": 3}},
					{"id": 30, "communicationType": "Fax", "value": "555-2", "defaultFlag": true, "type": {"id": 4}}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "customFields", Value: []any{}},
				{Op: "replace", Path: "firstName", Value: "Multiple"},
				{
					Op:   "add",
					Path: "/communicationItems/4",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "100"},
						Value:             "new.default@test.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
				},
				{
					Op:   "add",
					Path: "/communicationItems/4",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "200"},
						Value:             "555-9999",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
				},
				{Op: "remove", Path: "/communicationItems/3"},
				{Op: "remove", Path: "/communicationItems/2"},
				{Op: "remove", Path: "/communicationItems/1"},
				{Op: "remove", Path: "/communicationItems/0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Mock server to handle fetchContact and fetchMissingCommunicationItemIds.
			server := mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v4_6_release/apis/3.0/company/contacts/123"),
					Then: mockserver.ResponseString(http.StatusOK, tt.respContact),
				}, {
					If:   mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
					Then: mockserver.ResponseString(http.StatusOK, tt.respCommunicationItems),
				}},
			}.Server()
			defer server.Close()

			connector, err := constructTestConnector(server)
			if err != nil {
				t.Fatalf("failed to construct connector: %v", err)
			}

			ops, err := connector.contactsFullUpdatePayload(t.Context(), tt.payload, "123")
			if err != nil {
				t.Fatalf("contactsFullUpdatePayload failed: %v", err)
			}

			res := testutils.NewCompareResult()
			res.Assert("operations", tt.expected, ops)
			res.Validate(t, tt.name)
		})
	}
}

// TestContactsPartialUpdatePayload verifies that partial contact updates (PATCH) correctly
// translate virtual communication field operations into concrete ConnectWise JSON Patch operations.
func TestContactsPartialUpdatePayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		input                  []patchOperationPayload
		respContact            string
		respCommunicationItems string
		expected               []patchOperationPayload
	}{
		{
			name: "Non-communication operations only",
			input: []patchOperationPayload{
				{Op: "replace", Path: "/firstName", Value: "Jane"},
			},
			expected: []patchOperationPayload{
				{Op: "replace", Path: "/firstName", Value: "Jane"},
			},
		},
		{
			name: "Virtual communication field update (replace)",
			input: []patchOperationPayload{
				{Op: "replace", Path: "AMPERSAND-defaultEmail", Value: "new@example.com"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{
						"id": 10,
						"communicationType": "Email",
						"value": "old@example.com",
						"defaultFlag": true,
						"type": {"id": 1}
					}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "/communicationItems/0/value", Value: "new@example.com"},
			},
		},
		{
			name: "Virtual communication field update (add new type)",
			input: []patchOperationPayload{
				{Op: "replace", Path: "AMPERSAND-defaultPhone", Value: "555-1234"},
				{Op: "replace", Path: "AMPERSAND-defaultPhoneId", Value: "2"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": []
			}`,
			expected: []patchOperationPayload{
				{
					Op:   "add",
					Path: "/communicationItems/0",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "2"},
						Value:             "555-1234",
						DefaultFlag:       true,
						CommunicationType: "Phone",
					},
				},
			},
		},
		{
			name: "Virtual communication field removal",
			input: []patchOperationPayload{
				{Op: "remove", Path: "AMPERSAND-defaultEmail"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{
						"id": 10,
						"communicationType": "Email",
						"value": "old@example.com",
						"defaultFlag": true,
						"type": {"id": 1}
					}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "remove", Path: "/communicationItems/0"},
			},
		},
		{
			name: "Fetch missing IDs for partial update",
			input: []patchOperationPayload{
				{Op: "replace", Path: "AMPERSAND-defaultEmail", Value: "fetch@example.com"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": []
			}`,
			respCommunicationItems: `[
				{"id": 5, "emailFlag": true, "defaultFlag": true}
			]`,
			expected: []patchOperationPayload{
				{
					Op:   "add",
					Path: "/communicationItems/0",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "5"},
						Value:             "fetch@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
				},
			},
		},
		{
			name: "Mixed real and virtual fields",
			input: []patchOperationPayload{
				{Op: "replace", Path: "/lastName", Value: "Smith"},
				{Op: "replace", Path: "AMPERSAND-defaultEmail", Value: "smith@example.com"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{
						"id": 10,
						"communicationType": "Email",
						"value": "old@example.com",
						"defaultFlag": true,
						"type": {"id": 1}
					}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "/lastName", Value: "Smith"},
				{Op: "replace", Path: "/communicationItems/0/value", Value: "smith@example.com"},
			},
		},
		{
			name: "Remove multiple virtual fields in one PATCH",
			input: []patchOperationPayload{
				{Op: "remove", Path: "AMPERSAND-defaultEmail"},
				{Op: "remove", Path: "AMPERSAND-defaultFax"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{"id": 10, "communicationType": "Email", "value": "e@test.com", "defaultFlag": true, "type": {"id": 1}},
					{"id": 20, "communicationType": "Phone", "value": "555-1", "defaultFlag": true, "type": {"id": 2}},
					{"id": 30, "communicationType": "Fax", "value": "555-2", "defaultFlag": true, "type": {"id": 3}}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "remove", Path: "/communicationItems/2"},
				{Op: "remove", Path: "/communicationItems/0"},
			},
		},
		{
			name: "Change default type ID set to a new value",
			input: []patchOperationPayload{
				{Op: "replace", Path: "AMPERSAND-defaultEmail", Value: "new.default@test.com"},
				{Op: "replace", Path: "AMPERSAND-defaultEmailId", Value: "15"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{"id": 10, "communicationType": "Email", "value": "old.default@test.com", "defaultFlag": true, "type": {"id": 8}}
				]
			}`,
			expected: []patchOperationPayload{
				{
					Op:   "add",
					Path: "/communicationItems/1",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "15"},
						Value:             "new.default@test.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
				},
			},
		},
		{
			name: "Only default communication item is updated",
			input: []patchOperationPayload{
				{Op: "replace", Path: "AMPERSAND-defaultEmail", Value: "updated.default@test.com"},
			},
			respContact: `{
				"id": 123,
				"communicationItems": [
					{"id": 10, "communicationType": "Email", "value": "old.default@test.com", "defaultFlag": false, "type": {"id": 8}},
					{"id": 11, "communicationType": "Email", "value": "other@test.com", "defaultFlag": true, "type": {"id": 1}}
				]
			}`,
			expected: []patchOperationPayload{
				{Op: "replace", Path: "/communicationItems/1/value", Value: "updated.default@test.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Mock server to handle fetchContact and fetchMissingCommunicationItemIds.
			server := mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v4_6_release/apis/3.0/company/contacts/123"),
					Then: mockserver.ResponseString(http.StatusOK, tt.respContact),
				}, {
					If:   mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
					Then: mockserver.ResponseString(http.StatusOK, tt.respCommunicationItems),
				}},
			}.Server()
			defer server.Close()

			connector, err := constructTestConnector(server)
			if err != nil {
				t.Fatalf("failed to construct connector: %v", err)
			}

			ops, err := connector.contactsPartialUpdatePayload(t.Context(), tt.input, "123")
			if err != nil {
				t.Fatalf("contactsPartialUpdatePayload failed: %v", err)
			}

			res := testutils.NewCompareResult()
			res.Assert("operations", tt.expected, ops)
			res.Validate(t, tt.name)
		})
	}
}

// TestMakeJsonPatchOperations verifies the logic for generating JSON Patch operations
// for adding, updating, or removing communication items based on the existing items registry.
func TestMakeJsonPatchOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		identifier        string
		value             string
		isRemove          bool
		communicationType string
		items             []readCommunicationItem
		itemsRegistry     map[string]int
		expectedOps       []patchOperationPayload
		expectedRemoves   []patchOperationPayload
	}{
		{
			name:            "Empty identifier",
			identifier:      "",
			value:           "test@example.com",
			expectedOps:     nil,
			expectedRemoves: nil,
		},
		{
			name:              "Add new item (not in registry)",
			identifier:        "1",
			value:             "new@example.com",
			isRemove:          false,
			communicationType: "Email",
			items:             []readCommunicationItem{},
			itemsRegistry:     map[string]int{},
			expectedOps: []patchOperationPayload{
				{
					Op:   "add",
					Path: "/communicationItems/0",
					Value: createCommunicationItemPayload{
						Type:              communicationItemTypePayload{Id: "1"},
						Value:             "new@example.com",
						DefaultFlag:       true,
						CommunicationType: "Email",
					},
				},
			},
			expectedRemoves: nil,
		},
		{
			name:              "Remove existing item",
			identifier:        "1",
			isRemove:          true,
			communicationType: "Email",
			items: []readCommunicationItem{
				{
					Id:   10,
					Type: readCommunicationItemType{Id: "1"},
				},
			},
			itemsRegistry: map[string]int{"1": 0},
			expectedOps:   nil,
			expectedRemoves: []patchOperationPayload{
				{
					Op:          "remove",
					Path:        "/communicationItems/0",
					removeIndex: 0,
				},
			},
		},
		{
			name:              "Update existing item value (already default)",
			identifier:        "1",
			value:             "updated@example.com",
			isRemove:          false,
			communicationType: "Email",
			items: []readCommunicationItem{
				{
					Id:          10,
					DefaultFlag: true,
					Type:        readCommunicationItemType{Id: "1"},
				},
			},
			itemsRegistry: map[string]int{"1": 0},
			expectedOps: []patchOperationPayload{
				{
					Op:    "replace",
					Path:  "/communicationItems/0/value",
					Value: "updated@example.com",
				},
			},
			expectedRemoves: nil,
		},
		{
			name:              "Update existing item and set defaultFlag",
			identifier:        "1",
			value:             "updated@example.com",
			isRemove:          false,
			communicationType: "Email",
			items: []readCommunicationItem{
				{
					Id:          10,
					DefaultFlag: false,
					Type:        readCommunicationItemType{Id: "1"},
				},
			},
			itemsRegistry: map[string]int{"1": 0},
			expectedOps: []patchOperationPayload{
				{
					Op:    "replace",
					Path:  "/communicationItems/0/value",
					Value: "updated@example.com",
				},
				{
					Op:    "replace",
					Path:  "/communicationItems/0/defaultFlag",
					Value: true,
				},
			},
			expectedRemoves: nil,
		},
		{
			name:              "Remove item not in registry (no-op)",
			identifier:        "2",
			isRemove:          true,
			communicationType: "Email",
			itemsRegistry:     map[string]int{"1": 0},
			expectedOps:       nil,
			expectedRemoves:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ops, removes := makeJsonPatchOperations(
				tt.identifier, tt.value, tt.isRemove,
				tt.communicationType, tt.items, tt.itemsRegistry,
			)

			res := testutils.NewCompareResult()
			res.Assert("ops", tt.expectedOps, ops)
			res.Assert("removes", tt.expectedRemoves, removes)
			res.Validate(t, tt.name)
		})
	}
}
