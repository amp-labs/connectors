// nolint:dupl,varnamelen
package codec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/test/utils/mockutils"
)

func TestRawJSON(t *testing.T) { // nolint:funlen
	t.Parallel()

	type userData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name       string
		jsonInput  string
		wantData   userData
		wantRaw    map[string]any
		marshalOut string // expected re-marshaled JSON (order-insensitive check)
		modifier   func(user *RawJSON[userData])
	}{
		{
			name:      "Simple identity",
			jsonInput: `{"id":"123","name":"Alice"}`,
			wantData:  userData{ID: "123", Name: "Alice"},
			wantRaw: map[string]any{
				"id":   "123",
				"name": "Alice",
			},
			marshalOut: `{"id":"123","name":"Alice"}`,
		},
		{
			name:      "Additional fields not part of concrete struct",
			jsonInput: `{"id":"123","name":"Alice","age":18}`,
			wantData:  userData{ID: "123", Name: "Alice"},
			wantRaw: map[string]any{
				"id":   "123",
				"name": "Alice",
				"age":  18.0,
			},
			marshalOut: `{"id":"123","name":"Alice","age":18}`,
		},
		{
			name:      "Modifications to concrete struct are preserved",
			jsonInput: `{"id":"123","name":"Bob","age":33}`,
			wantData:  userData{ID: "123", Name: "Bob"},
			wantRaw: map[string]any{
				"id":   "123",
				"name": "Bob",
				"age":  33.0,
			},
			modifier: func(user *RawJSON[userData]) {
				// After changing Data.Name and re-marshaling,
				// it should override "name" in Raw.
				user.Data.Name = "New"
			},
			marshalOut: `{"id":"123","name":"New","age":33}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var user RawJSON[userData]

			if err := json.Unmarshal([]byte(tt.jsonInput), &user); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			// Check that the typed Data matches the expected value
			if !reflect.DeepEqual(user.Data, tt.wantData) {
				t.Errorf("Data mismatch:\n got: %#v\nwant: %#v", user.Data, tt.wantData)
			}

			// Check that the Raw map contains the expected key-value pairs
			if !reflect.DeepEqual(user.Raw, tt.wantRaw) {
				t.Errorf("Raw mismatch:\n got: %#v\nwant: %#v", user.Raw, tt.wantRaw)
			}

			// Apply any test-specific modifications to Data
			if tt.modifier != nil {
				tt.modifier(&user)
			}

			// Check that the marshaled output matches the expected JSON
			if !mockutils.JSONComparator.Equals(user, tt.marshalOut) {
				data, _ := user.MarshalJSON()
				t.Errorf("MarshalJSON mismatch:\n got: %s\nwant: %v", data, tt.marshalOut)
			}
		})
	}
}

func TestNestedRawJSON(t *testing.T) { // nolint:funlen
	t.Parallel()

	type userData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type Contact = RawJSON[userData]

	type companyData struct {
		CompanyName string   `json:"companyName"`
		Phone       string   `json:"phone"`
		Contact     *Contact `json:"contact"`
	}

	contactAlice, _ := NewRawJSON(userData{ID: "123", Name: "Alice"})

	tests := []struct {
		name       string
		jsonInput  string
		wantData   companyData
		wantRaw    map[string]any
		marshalOut string // expected re-marshaled JSON (order-insensitive check)
		modifier   func(user *RawJSON[companyData])
	}{
		{
			name: "Nested RawJSON with access to inner Raw",
			jsonInput: `{
				"companyName":"Nike",
				"contact":{"id":"123","name":"Alice"},
				"phone":"555-1234"
			}`,
			wantData: companyData{
				CompanyName: "Nike",
				Phone:       "555-1234",
				Contact:     contactAlice,
			},
			wantRaw: map[string]any{
				"companyName": "Nike",
				"phone":       "555-1234",
				"contact": map[string]any{
					"id":   "123",
					"name": "Alice",
				},
			},
			modifier: func(company *RawJSON[companyData]) {
				company.Data.Contact.Data.Name = "Bob" // New contact name
				company.Data.CompanyName = "Adidas"    // New company name
			},
			marshalOut: `{
				"companyName":"Adidas",
				"contact":{"id":"123","name":"Bob"},
				"phone":"555-1234"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var company RawJSON[companyData]

			if err := json.Unmarshal([]byte(tt.jsonInput), &company); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			// Check that the typed Data matches the expected value
			if !reflect.DeepEqual(company.Data, tt.wantData) {
				t.Errorf("Data mismatch:\n got: %#v\nwant: %#v", company.Data, tt.wantData)
			}

			// Check that the Raw map contains the expected key-value pairs
			if !reflect.DeepEqual(company.Raw, tt.wantRaw) {
				t.Errorf("Raw mismatch:\n got: %#v\nwant: %#v", company.Raw, tt.wantRaw)
			}

			// Apply any test-specific modifications to Data
			if tt.modifier != nil {
				tt.modifier(&company)
			}

			// Check that the marshaled output matches the expected JSON
			if !mockutils.JSONComparator.Equals(company, tt.marshalOut) {
				data, _ := company.MarshalJSON()
				t.Errorf("MarshalJSON mismatch:\n got: %s\nwant: %v", data, tt.marshalOut)
			}
		})
	}
}

func TestNewRawJSON(t *testing.T) {
	t.Parallel()

	type userData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	data := userData{ID: "1", Name: "Bob"}

	obj, err := NewRawJSON(data)
	if err != nil {
		t.Fatalf("NewRawJSON() error = %v", err)
	}

	if !reflect.DeepEqual(obj.Data, data) {
		t.Errorf("Data mismatch:\n got: %#v\nwant: %#v", obj.Data, data)
	}

	wantRaw := map[string]any{"id": "1", "name": "Bob"}
	if !reflect.DeepEqual(obj.Raw, wantRaw) {
		t.Errorf("Raw mismatch:\n got: %#v\nwant: %#v", obj.Raw, wantRaw)
	}
}
