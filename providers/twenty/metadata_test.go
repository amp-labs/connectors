package twenty

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	responseObjects := testutils.DataFromFile(t, "metadata-objects.json")
	responsePage1 := testutils.DataFromFile(t, "metadata-objects-page-1.json")
	responsePage2 := testutils.DataFromFile(t, "metadata-objects-page-2.json")
	responseUnauthenticated := testutils.DataFromFile(t, "error-unauthenticated.json")
	responseLegacyPage1 := legacyEnvelope(t, responsePage1)
	responseLegacyPage2 := legacyEnvelope(t, responsePage2)

	objectsServer := func() *mockserver.Conditional {
		return &mockserver.Conditional{
			Setup: mockserver.ContentJSON(),
			If: mockcond.And{
				mockcond.MethodGET(),
				mockcond.Path("/rest/metadata/objects"),
			},
			Then: mockserver.Response(http.StatusOK, responseObjects),
		}
	}

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown and inactive objects are reported per object",
			Input:      []string{"archivedThings", "invoices", "Companies"},
			Server:     objectsServer().Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"archivedThings": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
					"invoices":       mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
					"Companies":      mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:   "Custom object fields come from the workspace schema, with select values and join columns",
			Input:  []string{"vendorContracts"},
			Server: objectsServer().Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				// One-to-many relations are not in a depth=0 record, and the relation itself is carried
				// as its join column.
				"vendorContracts": {"timelineActivities", "attachments", "vendor"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"vendorContracts": {
						DisplayName: "Vendor contracts",
						Fields: map[string]common.FieldMetadata{
							"renewalNotes": {
								DisplayName:  "Renewal notes",
								ValueType:    common.ValueTypeString,
								ProviderType: "TEXT",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("e4e72b34-60be-41c5-a1f9-93ac28dcff7a"),
							},
							"billingAddress": {
								DisplayName:  "Billing address",
								ValueType:    common.ValueTypeOther,
								ProviderType: "ADDRESS",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("1fd36a32-5f4f-49c8-9921-ce104c0db451"),
							},
							"contractStatus": {
								DisplayName:  "Contract status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "SELECT",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("b478eaf4-eb18-4793-a682-03723bf2bcf6"),
								Values: []common.FieldValue{
									{Value: "DRAFT", DisplayValue: "Draft"},
									{Value: "SIGNED", DisplayValue: "Signed"},
									{Value: "EXPIRED", DisplayValue: "Expired"},
								},
							},
							"regions": {
								DisplayName:  "Regions",
								ValueType:    common.ValueTypeMultiSelect,
								ProviderType: "MULTI_SELECT",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("90bff50f-2cf1-47fa-b9fc-c27c5055e2b7"),
								Values: []common.FieldValue{
									{Value: "EUROPE", DisplayValue: "Europe"},
									{Value: "AMERICAS", DisplayValue: "Americas"},
								},
							},
							"vendorId": {
								DisplayName:  "Vendor",
								ValueType:    common.ValueTypeReference,
								ProviderType: "RELATION",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("92f18b1c-378c-4d91-955f-d184b9e5bc5c"),
							},
							"updatedAt": {
								DisplayName:  "Last update",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "DATE_TIME",
								ReadOnly:     new(true),
								IsRequired:   new(false),
								FieldId:      new("cd0f836d-a99f-44dc-b890-16ea695a84b5"),
							},
							"searchVector": {
								DisplayName:  "Search vector",
								ValueType:    common.ValueTypeOther,
								ProviderType: "TS_VECTOR",
								ReadOnly:     new(true),
								IsRequired:   new(false),
								FieldId:      new("16223cc8-37c8-4acd-8628-ae146e22bca9"),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:   "Standard object: inactive field left out, composite fields kept whole",
			Input:  []string{"companies"},
			Server: objectsServer().Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				"companies": {"legacyRating", "people", "vendorContracts"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"annualRevenue": {
								DisplayName:  "Annual Revenue",
								ValueType:    common.ValueTypeOther,
								ProviderType: "CURRENCY",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("5d004e0a-c78b-4ff0-8390-d1a679a2aa1c"),
							},
							"accountOwnerId": {
								DisplayName:  "Account Owner",
								ValueType:    common.ValueTypeReference,
								ProviderType: "RELATION",
								ReadOnly:     new(false),
								IsRequired:   new(false),
								FieldId:      new("257ea92c-9c39-49b6-8b90-a06c6c13716c"),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Every page of the object metadata is read",
			Input: []string{"people", "vendorContracts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/rest/metadata/objects"),
						mockcond.QueryParam("limit", "200"),
						mockcond.QueryParam("starting_after", "6762a14f-e73b-4ccb-a130-7641a1b20770"),
					},
					Then: mockserver.Response(http.StatusOK, responsePage2),
				}, {
					If: mockcond.And{
						mockcond.Path("/rest/metadata/objects"),
						mockcond.QueryParam("limit", "200"),
						mockcond.QueryParamsMissing("starting_after"),
					},
					Then: mockserver.Response(http.StatusOK, responsePage1),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"people": {
						DisplayName: "People",
						Fields:      map[string]common.FieldMetadata{},
					},
					"vendorContracts": {
						DisplayName: "Vendor contracts",
						Fields:      map[string]common.FieldMetadata{},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "The legacy envelope {data: {objects: [...]}} is read, every page",
			Input: []string{"people", "vendorContracts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/rest/metadata/objects"),
						mockcond.QueryParam("starting_after", "6762a14f-e73b-4ccb-a130-7641a1b20770"),
					},
					Then: mockserver.Response(http.StatusOK, responseLegacyPage2),
				}, {
					If: mockcond.And{
						mockcond.Path("/rest/metadata/objects"),
						mockcond.QueryParamsMissing("starting_after"),
					},
					Then: mockserver.Response(http.StatusOK, responseLegacyPage1),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"people":          {DisplayName: "People", Fields: map[string]common.FieldMetadata{}},
					"vendorContracts": {DisplayName: "Vendor contracts", Fields: map[string]common.FieldMetadata{}},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Any other envelope is an error",
			Input: []string{"people"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, []byte(`{"data":{"items":[]},"pageInfo":{"hasNextPage":false}}`)),
			}.Server(),
			ExpectedErrs: []error{errMetadataEnvelope},
		},
		{
			Name:  "A next page without a new cursor fails instead of looping",
			Input: []string{"people"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responsePage1),
			}.Server(),
			ExpectedErrs: []error{common.ErrNextPageInvalid},
		},
		{
			Name:  "A next page without any cursor fails instead of stopping short",
			Input: []string{"people"},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, []byte(`{"data":[],"totalCount":44,`+
					`"pageInfo":{"hasNextPage":true,"hasPreviousPage":false,"startCursor":null,"endCursor":null}}`)),
			}.Server(),
			ExpectedErrs: []error{common.ErrNextPageInvalid},
		},
		{
			Name:  "A 404 page that is not JSON (wrong domain, a proxy in front) is not retried",
			Input: []string{"companies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.ResponseString(http.StatusNotFound, "<html><body>404 Not Found</body></html>"),
			}.Server(),
			ExpectedErrs: []error{common.ErrBadRequest, common.ErrNotFound},
		},
		{
			Name:  "A rejected key is an auth error carrying Twenty's message",
			Input: []string{"companies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseUnauthenticated),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("Token invalid."),
			},
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

// legacyEnvelope serves a metadata page as {"data": {"objects": [...]}}, the shape Twenty v2.50.0 sends when
// IS_REST_METADATA_API_NEW_FORMAT_DIRECT is off (seen live by switching the flag); the objects and pageInfo
// are those of the page given.
func legacyEnvelope(t *testing.T, page []byte) []byte {
	t.Helper()

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(page, &envelope); err != nil {
		t.Fatal(err)
	}

	objects, err := json.Marshal(map[string]json.RawMessage{"objects": envelope["data"]})
	if err != nil {
		t.Fatal(err)
	}

	envelope["data"] = objects

	legacy, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}

	return legacy
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "crm.example.com",
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}

// A field without writability (an attribute v2.50.0 sends, other versions may not) is writable; only an
// explicit non-OPEN writability or a UI-locked system field is read-only.
func TestFieldReadOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		field fieldMetadata
		want  bool
	}{
		{"open field", fieldMetadata{Writability: "OPEN"}, false},
		{"writability missing", fieldMetadata{}, false},
		{"system writability", fieldMetadata{Writability: "SYSTEM"}, true},
		{"system field locked in the UI", fieldMetadata{IsSystem: true, IsUIReadOnly: true, Writability: "OPEN"}, true},
		{"system field editable in the UI", fieldMetadata{IsSystem: true, Writability: "OPEN"}, false},
	}

	for _, tt := range tests {
		if got := tt.field.readOnly(); got != tt.want {
			t.Errorf("%s: readOnly() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
