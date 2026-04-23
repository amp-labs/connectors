package gusto

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
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
			Comparator:   testroutines.ComparatorSubsetMetadata,
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
			Comparator:   testroutines.ComparatorSubsetMetadata,
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
			Comparator:   testroutines.ComparatorSubsetMetadata,
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
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
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
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
