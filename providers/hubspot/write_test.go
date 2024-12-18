package hubspot

import (
	"errors"
	"reflect"
	"testing"
)

func TestFormatData(t *testing.T) { //nolint:funlen
	t.Parallel()

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
					"hs_timestamp":    1734527635844,
					"hs_email_status": "SENT",
					"hs_email_headers": "" +
						"{'from':{'email':null},'to':[{'email':null,'firstName':'Some','lastName':'Person'}]}",
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
					"hs_timestamp":    1734527635844,
					"hs_email_status": "SENT",
					"hs_email_headers": "" +
						"{'from':{'email':null},'to':[{'email':null,'firstName':'Some','lastName':'Person'}]}",
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

	for _, ttc := range tests {
		t.Run(ttc.name, func(t *testing.T) {
			t.Parallel()

			got, err := formatData(ttc.input)
			if (err != nil) != ttc.wantErr {
				t.Fatalf("formatData() error = %v, wantErr %v", err, ttc.wantErr)
			}

			if ttc.wantErr {
				if !errors.Is(err, ttc.expectedErr) {
					t.Fatalf("formatData() error = %v, expectedErr %v", err, ttc.expectedErr)
				}

				return
			}

			if !reflect.DeepEqual(got, ttc.want) {
				t.Errorf("formatData() = %#v, want %#v", got, ttc.want)
			}
		})
	}
}
