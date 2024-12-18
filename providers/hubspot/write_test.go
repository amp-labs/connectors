package hubspot

import (
	"errors"
	"reflect"
	"testing"
)

func TestFormatData(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		want        map[string]any
		wantErr     bool
		expectedErr error
	}{
		{
			name: "invalid data format (not a map[string]any)",
			input: map[int]string{
				1: "value",
			},
			wantErr:     true,
			expectedErr: ErrInvalidDataFormat,
		},
		{
			name: "already has properties field",
			input: map[string]any{
				string(ObjectFieldProperties): map[string]any{"foo": "bar", "hs_object_id": "123"},
			},
			want: map[string]any{
				string(ObjectFieldProperties): map[string]any{"foo": "bar", "hs_object_id": "123"},
			},
		},
		{
			name: "already has associations field",
			input: map[string]any{
				string(ObjectFieldAssociations): []any{"some_association"},
			},
			want: map[string]any{
				string(ObjectFieldAssociations): []any{"some_association"},
			},
		},
		{
			name: "no properties or associations, wrap data in properties",
			input: map[string]any{
				"somekey": "somevalue",
			},
			want: map[string]any{
				string(ObjectFieldProperties): map[string]any{
					"somekey": "somevalue",
				},
			},
		},
		{
			name: "properties and associations, return as is",
			input: map[string]any{
				"properties": map[string]any{
					"hs_timestamp":     1734527635844,
					"hs_email_status":  "SENT",
					"hs_email_headers": "{\"from\":{\"email\":null},\"to\":[{\"email\":null,\"firstName\":\"Some\",\"lastName\":\"Person\"}]}",
				},
				"associations": []map[string]any{
					{
						"to": map[string]any{
							"id": "85861910068",
						},
					},
					{
						"types": []map[string]any{
							{
								"associationCategory": "HUBSPOT_DEFINED",
								"associationTypeId":   11,
							},
						},
					},
				},
			},
			want: map[string]any{
				"properties": map[string]any{
					"hs_timestamp":     1734527635844,
					"hs_email_status":  "SENT",
					"hs_email_headers": "{\"from\":{\"email\":null},\"to\":[{\"email\":null,\"firstName\":\"Some\",\"lastName\":\"Person\"}]}",
				},
				"associations": []map[string]any{
					{
						"to": map[string]any{
							"id": "85861910068",
						},
					},
					{
						"types": []map[string]any{
							{
								"associationCategory": "HUBSPOT_DEFINED",
								"associationTypeId":   11,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // pin variable
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatData(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("formatData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("formatData() error = %v, expectedErr %v", err, tt.expectedErr)
				}

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatData() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
