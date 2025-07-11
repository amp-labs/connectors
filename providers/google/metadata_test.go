package google

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for CalendarList and Settings",
			Input:      []string{"calendarList", "settings", "events"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"calendarList": {
						DisplayName: "Calendars",
						Fields: map[string]common.FieldMetadata{
							"backgroundColor": {
								DisplayName:  "Background Color",
								ValueType:    "string",
								ProviderType: "string",
							},
							"notificationSettings": {
								DisplayName:  "Notification Settings",
								ValueType:    "other",
								ProviderType: "object",
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
					"events": {
						DisplayName: "Events",
						Fields: map[string]common.FieldMetadata{
							"summary": {
								DisplayName:  "Summary",
								ValueType:    "string",
								ProviderType: "string",
							},
							"privateCopy": {
								DisplayName:  "Private Copy",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestCalendarConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestCalendarConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              providers.ModuleGoogleCalendar,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
