package granola

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

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()
	notesResponse := testutils.DataFromFile(t, "notes.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe notes object by sampling first record from data array",
			Input: []string{"notes"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "1"),
				},
				Then: mockserver.Response(http.StatusOK, notesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"notes": {
						DisplayName: "Notes",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"object": {
								DisplayName:  "object",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"title": {
								DisplayName:  "title",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"owner": {
								DisplayName:  "owner",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when server responds with 500",
			Input: []string{"notes"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "1"),
				},
				Then: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"notes": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when response is empty",
			Input: []string{"notes"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "1"),
				},
				Then: mockserver.Response(http.StatusOK, []byte("{}")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"notes": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when notes array is empty",
			Input: []string{"notes"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v0/notes"),
					mockcond.QueryParam("page_size", "1"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"notes": []}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"notes": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
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
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
