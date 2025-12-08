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

func TestCalendarListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
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
								Values:       nil,
							},
							"value": {
								DisplayName:  "Value",
								ValueType:    "string",
								ProviderType: "string",
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

func TestContactsListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for CalendarList and Settings",
			Input:      []string{"myConnections", "peopleDirectory"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"myConnections": {
						DisplayName: "Connections",
						Fields: map[string]common.FieldMetadata{
							"phoneNumbers": {
								DisplayName:  "Phone Numbers",
								ValueType:    "other",
								ProviderType: "array",
							},
							"nicknames": {
								DisplayName:  "Nicknames",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
					"peopleDirectory": {
						DisplayName: "People Directory",
						Fields: map[string]common.FieldMetadata{
							"birthdays": {
								DisplayName:  "Birthdays",
								ValueType:    "other",
								ProviderType: "array",
							},
							"names": {
								DisplayName:  "Names",
								ValueType:    "other",
								ProviderType: "array",
							},
							"skills": {
								DisplayName:  "Skills",
								ValueType:    "other",
								ProviderType: "array",
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
				return constructTestContactsConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestCalendarConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleGoogleCalendar)
}

func constructTestContactsConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleGoogleContacts)
}

func constructTestConnector(serverURL string, moduleID common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              moduleID,
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
