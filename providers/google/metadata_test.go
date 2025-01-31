package google

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"calendarList", "settings"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"calendarList": {
						DisplayName: "Calendars",
						Fields: map[string]common.FieldMetadata{
							"description": {
								DisplayName:  "Description",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"primary": {
								DisplayName:  "Primary",
								ValueType:    "boolean",
								ProviderType: "boolean",
								ReadOnly:     false,
								Values:       nil,
							},
							"defaultReminders": {
								DisplayName:  "Default Reminders",
								ValueType:    "other",
								ProviderType: "array",
								ReadOnly:     false,
								Values:       nil,
							},
						},
					},
					"settings": {
						DisplayName: "Settings",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"value": {
								DisplayName:  "Value",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
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
