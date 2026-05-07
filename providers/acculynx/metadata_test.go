package acculynx

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object returns ErrObjectNotSupported",
			Input:      []string{"nonexistent"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:   "Successfully describe top-level jobs object",
			Input:  []string{"jobs"},
			Server: customFieldDefinitionsServer(customFieldDefinitionsEmptyResponse),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						FieldsMap: map[string]string{
							"id":               "id",
							"createdDate":      "createdDate",
							"currentMilestone": "currentMilestone",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:   "Successfully describe top-level contacts object",
			Input:  []string{"contacts"},
			Server: customFieldDefinitionsServer(customFieldDefinitionsEmptyResponse),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							"id":        "id",
							"firstName": "firstName",
							"lastName":  "lastName",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:   "Slash-named nested object resolves correctly",
			Input:  []string{"jobs/contacts"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs/contacts": {
						DisplayName: "Job Contacts",
						FieldsMap: map[string]string{
							"id": "id",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:  "Contacts metadata enriched with custom field definitions",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsFixture),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							// Built-in fields (subset of the static schema).
							"id":        "id",
							"firstName": "firstName",
							// customer_preference and preferred_contact_method
							// are custom fields, slugs derived from their
							// labels — display names map back to the labels.
							"customer_preference":      "Customer Preference",
							"preferred_contact_method": "Preferred Contact Method",
						},
					},
				},
			},
		},
		{
			Name:  "Jobs metadata enriched with custom field definitions",
			Input: []string{"jobs"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsFixture),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						FieldsMap: map[string]string{
							"id": "id",
							// estimated_squares is a custom field on jobs —
							// proves entityType bucketing works (this
							// definition has entityType=job in the fixture).
							"estimated_squares": "Estimated Squares",
						},
					},
				},
			},
		},
		{
			Name:   "Successfully describe multiple objects at once",
			Input:  []string{"jobs", "users", "calendars"},
			Server: customFieldDefinitionsServer(customFieldDefinitionsEmptyResponse),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						FieldsMap: map[string]string{
							"id": "id",
						},
					},
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":          "id",
							"displayName": "displayName",
							"email":       "email",
						},
					},
					"calendars": {
						DisplayName: "Calendars",
						FieldsMap: map[string]string{
							"id":   "id",
							"name": "name",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}

// customFieldDefinitionsServer returns a mock server that responds to the
// /company-settings/custom-fields endpoint with the given fixture, and 500s
// on anything else. ListObjectMetadata is strict on definitions-fetch
// failure, so every test that targets contacts/jobs needs this mock — even
// tests that don't otherwise care about custom fields.
func customFieldDefinitionsServer(fixture []byte) *httptest.Server {
	return mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/company-settings/custom-fields"),
				},
				Then: mockserver.Response(http.StatusOK, fixture),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
}
