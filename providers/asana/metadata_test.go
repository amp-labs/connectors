package asana

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	allocationresponse := testutils.DataFromFile(t, "allocations.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"allocations"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/api/1.0/allocations"),
					Then: mockserver.Response(http.StatusOK, allocationresponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"allocations": {
						DisplayName: "allocations",
						FieldsMap: map[string]string{
							"gid":              "gid",
							"resource_type":    "resource_type",
							"start_date":       "start_date",
							"end_date":         "end_date",
							"effort":           "effort",
							"assignee":         "assignee",
							"created_by":       "created_by",
							"parent":           "parent",
							"resource_subtype": "resource_subtype",
						},
					},
				},
				Errors: nil,
			},
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
		WithAuthenticatedClient(http.DefaultClient),
	)

	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(serverURL)

	return connector, nil
}
