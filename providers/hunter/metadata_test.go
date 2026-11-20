package hunter

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	leadsResponse := testutils.DataFromFile(t, "leads-read.json")
	campaignsResponse := testutils.DataFromFile(t, "campaigns-read.json")
	emptyResponse := testutils.DataFromFile(t, "empty-read.json")
	accountResponse := testutils.DataFromFile(t, "account-read.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"leads", "campaigns"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v2/leads"),
					Then: mockserver.Response(http.StatusOK, leadsResponse),
				}, {
					If:   mockcond.Path("/v2/campaigns"),
					Then: mockserver.Response(http.StatusOK, campaignsResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"leads": {
						DisplayName: "Leads",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeFloat,
							},
							"email": {
								DisplayName: "email",
								ValueType:   common.ValueTypeString,
							},
							"phone_number": {
								DisplayName: "phone_number",
								ValueType:   common.ValueTypeOther,
							},
							"leads_list": {
								DisplayName: "leads_list",
								ValueType:   common.ValueTypeOther,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"email":        "email",
							"phone_number": "phone_number",
							"leads_list":   "leads_list",
						},
					},
					"campaigns": {
						DisplayName: "Campaigns",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName: "name",
								ValueType:   common.ValueTypeString,
							},
							"recipients_count": {
								DisplayName: "recipients_count",
								ValueType:   common.ValueTypeFloat,
							},
							"archived": {
								DisplayName: "archived",
								ValueType:   common.ValueTypeBoolean,
							},
							"owner": {
								DisplayName: "owner",
								ValueType:   common.ValueTypeOther,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe singleton object",
			Input: []string{"account"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v2/account"),
					Then: mockserver.Response(http.StatusOK, accountResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account": {
						DisplayName: "Account",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName: "email",
								ValueType:   common.ValueTypeString,
							},
							"plan_level": {
								DisplayName: "plan_level",
								ValueType:   common.ValueTypeFloat,
							},
							"requests": {
								DisplayName: "requests",
								ValueType:   common.ValueTypeOther,
							},
						},
						FieldsMap: map[string]string{
							"email":      "email",
							"plan_level": "plan_level",
							"requests":   "requests",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Object with no records cannot be sampled",
			Input: []string{"leads_lists"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v2/leads_lists"),
					Then: mockserver.Response(http.StatusOK, emptyResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"leads_lists": common.ErrMissingExpectedValues,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
