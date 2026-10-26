package reply

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	responseCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions.json")
	responseCollidingCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions-colliding.json")
	responseAccountCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions-accounts.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:  "Static metadata without custom fields makes no API call",
			Input: []string{"tasks", "inbox/threads"},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				// Any request would fail the test: these objects resolve from
				// the embedded schemas alone.
				Always: mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tasks": {
						DisplayName: "Tasks",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"taskType": {
								DisplayName:  "taskType",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{Value: "toDo", DisplayValue: "toDo"},
									{Value: "call", DisplayValue: "call"},
									{Value: "meeting", DisplayValue: "meeting"},
									{Value: "linkedIn", DisplayValue: "linkedIn"},
									{Value: "manualEmail", DisplayValue: "manualEmail"},
									{Value: "sms", DisplayValue: "sms"},
									{Value: "whatsApp", DisplayValue: "whatsApp"},
								},
							},
						},
					},
					"inbox/threads": {
						DisplayName: "Inbox Threads",
						Fields: map[string]common.FieldMetadata{
							"subject": {
								DisplayName:  "subject",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Custom fields are merged into Contacts metadata by title",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							// Deal Stage is a custom field: it comes from /v3/custom-fields, not from the static schema.
							"Deal Stage": {
								DisplayName:  "Deal Stage",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     new(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Account custom fields are merged into Contact Accounts metadata by title",
			Input: []string{"contact-accounts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/custom-fields"),
					mockcond.QueryParam("entityType", "account"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountCustomFields),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contact-accounts": {
						DisplayName: "Contact Accounts",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							// Renewal Date is a custom account field.
							"Renewal Date": {
								DisplayName:  "Renewal Date",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     new(true),
							},
							// Seats is a custom account field of type number.
							"Seats": {
								DisplayName:  "Seats",
								ValueType:    "float",
								ProviderType: "number",
								IsCustom:     new(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Contacts and Contact Accounts each load the definitions of their own entity type",
			Input: []string{"contacts", "contact-accounts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v3/custom-fields"),
						mockcond.QueryParam("entityType", "contact"),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.Path("/v3/custom-fields"),
						mockcond.QueryParam("entityType", "account"),
					},
					Then: mockserver.Response(http.StatusOK, responseAccountCustomFields),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				"contacts":         {"Renewal Date", "Seats"},
				"contact-accounts": {"Deal Stage"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							// Deal Stage is a custom field: it comes from /v3/custom-fields, not from the static schema.
							"Deal Stage": {
								DisplayName:  "Deal Stage",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     new(true),
							},
						},
					},
					"contact-accounts": {
						DisplayName: "Contact Accounts",
						Fields: map[string]common.FieldMetadata{
							// Renewal Date is a custom account field.
							"Renewal Date": {
								DisplayName:  "Renewal Date",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     new(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Custom field titled like a built-in field does not replace it",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, responseCollidingCustomFields),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				"contacts": {"Industry"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"industry": {
								DisplayName:  "industry",
								ValueType:    "string",
								ProviderType: "string",
							},
							// Deal Stage is a custom field: it comes from /v3/custom-fields, not from the static schema.
							"Deal Stage": {
								DisplayName:  "Deal Stage",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     new(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Team without account custom fields still gets Contact Accounts metadata",
			Input: []string{"contact-accounts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusForbidden, nil),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contact-accounts": {
						DisplayName: "Contact Accounts",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Forbidden contact custom fields fail Contacts metadata",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusForbidden, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrResolvingCustomFields, common.ErrForbidden},
		},
		{
			Name:  "Custom field endpoint failure fails Contacts metadata",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrResolvingCustomFields, common.ErrServer},
		},
		{
			Name:       "Unknown object is reported in the errors map",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
