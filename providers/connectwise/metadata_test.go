package connectwise

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactSample := testutils.DataFromFile(t, "read/one-contact.json")
	responseCustomFields := testutils.DataFromFile(t, "custom-fields/definitions.json")
	responseCommunicationItems1 := testutils.DataFromFile(t, "metadata/communication-items/1-first-page.json")
	responseCommunicationItems2 := testutils.DataFromFile(t, "metadata/communication-items/2-last-page.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:  "Successfully describe companies with metadata",
			Input: []string{"companies"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v4_6_release/apis/3.0/company/companies"),
					mockcond.QueryParam("pageSize", "1"),
					mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
				},
				Then: mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"city": {
								DisplayName:  "city",
								ValueType:    "string",
								ProviderType: "string",
							},
							"taxCode": {
								DisplayName:  "taxCode",
								ValueType:    "other",
								ProviderType: "object",
							},
						},
					},
				},
			},
		},
		{
			Name:  "Contacts with custom metadata",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Sample one record of Contacts.
					If: mockcond.And{
						mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
						mockcond.QueryParam("pageSize", "1"),
						mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseContactSample),
				}, {
					// Query custom fields by ids.
					If: mockcond.And{
						mockcond.Path("/v4_6_release/apis/3.0/system/userDefinedFields"),
						conditionsParamHasIds(
							"15", "16", "51", "52", "53", "54", "55", "56", "59", "60", "61", "62",
							"63", "64", "65", "66", "67", "76", "77", "78", "79", "80", "83", "88",
						),
						mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					// List communication items.
					If: mockcond.And{
						mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
						mockcond.QueryParamsMissing("page"),
						mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
					},
					Then: mockserver.ResponseChainedFuncs(
						mockserver.Header("Link", `<`+testconn.URLTestServer+
							`/v4_6_release/apis/3.0/company/communicationTypes?page=2>; rel="next"`,
						),
						mockserver.Response(http.StatusOK, responseCommunicationItems1),
					),
				}, {
					If: mockcond.And{
						mockcond.Path("/v4_6_release/apis/3.0/company/communicationTypes"),
						mockcond.QueryParam("page", "2"),
						mockcond.Header(http.Header{"ClientId": []string{"dummy-client-id"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCommunicationItems2),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "string",
								ProviderType: "string",
							},
							"gender": {
								DisplayName:  "gender",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{{
									Value:        "Female",
									DisplayValue: "Female",
								}, {
									Value:        "Male",
									DisplayValue: "Male",
								}},
							},
							"customField80": {
								DisplayName:  "Mobile Phone",
								ValueType:    "string",
								ProviderType: "EntryField_Text",
								ReadOnly:     new(false),
								IsCustom:     new(true),
								IsRequired:   new(false),
							},
							"customField16": {
								DisplayName:  "is_synced",
								ValueType:    "multiSelect",
								ProviderType: "List_Text",
								ReadOnly:     new(false),
								IsCustom:     new(true),
								IsRequired:   new(false),
								Values: []common.FieldValue{{
									Value:        "true",
									DisplayValue: "true",
								}, {
									Value:        "false",
									DisplayValue: "false",
								}},
							},
							"customField83": {
								DisplayName:  "Hobby",
								ValueType:    "multiSelect",
								ProviderType: "List_Text",
								ReadOnly:     new(false),
								IsCustom:     new(true),
								IsRequired:   new(false),
								Values: []common.FieldValue{{
									Value:        "Traveling",
									DisplayValue: "Traveling",
								}, {
									Value:        "Swimming",
									DisplayValue: "Swimming",
								}, {
									Value:        "Skiing",
									DisplayValue: "Skiing",
								}, {
									Value:        "Hiking",
									DisplayValue: "Hiking",
								}},
							},
							"customField88": {
								DisplayName:  "Chooser",
								ValueType:    "singleSelect",
								ProviderType: "Option_Text",
								ReadOnly:     new(false),
								IsCustom:     new(true),
								IsRequired:   new(false),
								Values: []common.FieldValue{{
									Value:        "Party1",
									DisplayValue: "Party1",
								}, {
									Value:        "Party2",
									DisplayValue: "Party2",
								}, {
									Value:        "Party3",
									DisplayValue: "Party3",
								}},
							},
							"AMPERSAND-email1": {
								DisplayName:  "Email - Business",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-email2": {
								DisplayName:  "Email - Private",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-phone3": {
								DisplayName:  "Phone - Work",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-phone4": {
								DisplayName:  "Phone - Home",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-phone5": {
								DisplayName:  "Phone - Landline",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-fax6": {
								DisplayName:  "Fax - Work",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-fax7": {
								DisplayName:  "Fax - Home",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-email-default": {
								DisplayName:  "Default communication item for Email",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-phone-default": {
								DisplayName:  "Default communication item for Phone",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
							"AMPERSAND-fax-default": {
								DisplayName:  "Default communication item for Fax",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     new(false),
								IsCustom:     new(false),
								IsRequired:   new(false),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: server.Client(),
			Metadata: map[string]string{
				"clientId": "dummy-client-id",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}

func conditionsParamHasIds(expectedIds ...string) mockcond.CustomCondition {
	return func(w http.ResponseWriter, r *http.Request) bool {
		query := r.URL.Query()
		param := query.Get("conditions")
		if param == "" {
			return false
		}

		re := regexp.MustCompile(`id in \(((?:[0-9]+,)*[0-9]+)\)`)
		m := re.FindStringSubmatch(param)
		if len(m) < 2 {
			return false
		}

		parts := strings.Split(m[1], ",")
		actual := make([]string, 0, len(parts))
		for _, p := range parts {
			actual = append(actual, p)
		}

		sort.Strings(actual)
		sort.Strings(expectedIds)

		return reflect.DeepEqual(actual, expectedIds)
	}
}
