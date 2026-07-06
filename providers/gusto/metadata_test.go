package gusto

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:   "Successfully describe employees object",
			Input:  []string{"employees"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"employees": {
						DisplayName: "Employees",
						FieldsMap: map[string]string{
							"uuid":       "uuid",
							"first_name": "first_name",
							"last_name":  "last_name",
							"email":      "email",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testconn.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe companies object",
			Input:  []string{"companies"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "Companies",
						FieldsMap: map[string]string{
							"uuid": "uuid",
							"name": "name",
							"ein":  "ein",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testconn.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe locations object",
			Input:  []string{"locations"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"locations": {
						DisplayName: "Locations",
						FieldsMap: map[string]string{
							"uuid":   "uuid",
							"city":   "city",
							"state":  "state",
							"active": "active",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testconn.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects",
			Input:  []string{"employees", "departments", "payrolls"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"employees": {
						DisplayName: "Employees",
						FieldsMap: map[string]string{
							"uuid":       "uuid",
							"first_name": "first_name",
						},
					},
					"departments": {
						DisplayName: "Departments",
						FieldsMap: map[string]string{
							"uuid":  "uuid",
							"title": "title",
						},
					},
					"payrolls": {
						DisplayName: "Payrolls",
						FieldsMap: map[string]string{
							"uuid":      "uuid",
							"processed": "processed",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testconn.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
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
			Module:              common.ModuleRoot,
			AuthenticatedClient: server.Client(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(server.URL)

	return connector, nil
}
