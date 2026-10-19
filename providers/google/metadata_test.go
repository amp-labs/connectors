package google

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestCalendarListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseExtendedProperties := testutils.DataFromFile(t, "calendar/metadata/events-extended-properties.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Successful metadata for CalendarList and Settings",
			Input:      []string{"calendarList", "settings", "events"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
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
							"start": {
								DisplayName:  "Start",
								ValueType:    "other",
								ProviderType: "object",
							},
							"start.dateTime": {
								DisplayName:  "Start Date Time",
								ValueType:    "datetime",
								ProviderType: "string",
							},
							"end.date": {
								DisplayName:  "End Date",
								ValueType:    "date",
								ProviderType: "string",
							},
							"source.url": {
								DisplayName:  "Source Url",
								ValueType:    "string",
								ProviderType: "string",
							},
							"extendedProperties.private": {
								DisplayName:  "Extended Properties Private",
								ValueType:    "other",
								ProviderType: "object",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Events metadata discovers extendedProperties keys from sampled events",
			Input: []string{"events"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/calendar/v3/calendars/primary/events"),
					mockcond.QueryParam("maxResults", "2500"),
					mockcond.QueryParam("fields", "items(extendedProperties)"),
				},
				Then: mockserver.Response(http.StatusOK, responseExtendedProperties),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"events": {
						DisplayName: "Events",
						Fields: map[string]common.FieldMetadata{
							"extendedProperties.private": {
								DisplayName:  "Extended Properties Private",
								ValueType:    "other",
								ProviderType: "object",
							},
							"extendedProperties.private.crmId": {
								DisplayName:  "Extended Properties Private crmId",
								ValueType:    "string",
								ProviderType: "string",
								IsCustom:     new(true),
							},
							"extendedProperties.private.dealStage": {
								DisplayName:  "Extended Properties Private dealStage",
								ValueType:    "string",
								ProviderType: "string",
								IsCustom:     new(true),
							},
							"extendedProperties.shared.team": {
								DisplayName:  "Extended Properties Shared team",
								ValueType:    "string",
								ProviderType: "string",
								IsCustom:     new(true),
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestCalendarConnector(tt.Server.URL)
			})
		})
	}
}

func TestContactsListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Successful metadata for CalendarList and Settings",
			Input:      []string{"myConnections", "peopleDirectory"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestContactsConnector(tt.Server.URL)
			})
		})
	}
}

func TestMailListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Successful metadata for messages and drafts",
			Input:      []string{"messages", "drafts"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"messages": {
						DisplayName: "Messages",
						Fields: map[string]common.FieldMetadata{
							"threadId": {
								DisplayName:  "Thread Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     nil,
							},
						},
					},
					"drafts": {
						DisplayName: "Drafts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
							},
							"message": {
								DisplayName:  "Message",
								ValueType:    "other",
								ProviderType: "object",
								ReadOnly:     new(false),
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestMailConnector(tt.Server.URL)
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

func constructTestMailConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleGoogleGmail)
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
