package outplay

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	prospectAccountResponse := testutils.DataFromFile(t, "prospectaccount-read.json")
	callAnalysisResponse := testutils.DataFromFile(t, "callanalysis-read.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"prospectaccount", "callanalysis"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/v1/prospectaccount/search"),
					},
					Then: mockserver.Response(http.StatusOK, prospectAccountResponse),
				}, {
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/v1/callanalysis/list"),
					},
					Then: mockserver.Response(http.StatusOK, callAnalysisResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"prospectaccount": { //nolint:dupl
						DisplayName: "Prospectaccount",
						Fields: map[string]common.FieldMetadata{
							"accountid": {
								DisplayName:  "accountid",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"employeecount": {
								DisplayName:  "employeecount",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"industrytype": {
								DisplayName:  "industrytype",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"linkedin": {
								DisplayName:  "linkedin",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"twitter": {
								DisplayName:  "twitter",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"foundedyear": {
								DisplayName:  "foundedyear",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"city": {
								DisplayName:  "city",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"website": {
								DisplayName:  "website",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"stage": {
								DisplayName:  "stage",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"owner": {
								DisplayName:  "owner",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"fields": {
								DisplayName:  "fields",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
					"callanalysis": {
						DisplayName: "Callanalysis",
						Fields: map[string]common.FieldMetadata{
							"callmetadataid": {
								DisplayName:  "callmetadataid",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"callstarttime": {
								DisplayName:  "callstarttime",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"callduration": {
								DisplayName:  "callduration",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"recordingfilepath": {
								DisplayName:  "recordingfilepath",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"mimetype": {
								DisplayName:  "mimetype",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"meetingtype": {
								DisplayName:  "meetingtype",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"callsource": {
								DisplayName:  "callsource",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"createddate": {
								DisplayName:  "createddate",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"attendees": {
								DisplayName:  "attendees",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"unknownattendees": {
								DisplayName:  "unknownattendees",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
				},
			},
		},
		{
			Name:  "Handle single object metadata",
			Input: []string{"prospectaccount"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/v1/prospectaccount/search"),
					},
					Then: mockserver.Response(http.StatusOK, prospectAccountResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"prospectaccount": { //nolint:dupl
						DisplayName: "Prospectaccount",
						Fields: map[string]common.FieldMetadata{
							"accountid": {
								DisplayName:  "accountid",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"employeecount": {
								DisplayName:  "employeecount",
								ValueType:    "float",
								ProviderType: "",
								Values:       nil,
							},
							"industrytype": {
								DisplayName:  "industrytype",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"linkedin": {
								DisplayName:  "linkedin",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"twitter": {
								DisplayName:  "twitter",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"foundedyear": {
								DisplayName:  "foundedyear",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"city": {
								DisplayName:  "city",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"website": {
								DisplayName:  "website",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"stage": {
								DisplayName:  "stage",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"owner": {
								DisplayName:  "owner",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"fields": {
								DisplayName:  "fields",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
					},
				},
			},
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
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test",
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
