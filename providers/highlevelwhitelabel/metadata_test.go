package highlevelwhitelabel

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

	businessesResponse := testutils.DataFromFile(t, "businesses.json")
	calendarsGroupsResponse := testutils.DataFromFile(t, "calendars_groups.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"businesses", "calendars/groups"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/businesses/"),
						mockcond.QueryParam("locationId", "iV1BEzddaWWLqU2kXhcN"),
					},
					Then: mockserver.Response(http.StatusOK, businessesResponse),
				}, {
					If: mockcond.And{
						mockcond.Path("/calendars/groups"),
						mockcond.QueryParam("locationId", "iV1BEzddaWWLqU2kXhcN"),
					},
					Then: mockserver.Response(http.StatusOK, calendarsGroupsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"businesses": {
						DisplayName: "Businesses",
						Fields: map[string]common.FieldMetadata{
							"customFields": {
								DisplayName: "customFields",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"locationId": {
								DisplayName: "locationId",
								ValueType:   "other",
							},
							"createdBy": {
								DisplayName: "createdBy",
								ValueType:   "other",
							},
							"createdAt": {
								DisplayName: "createdAt",
								ValueType:   "other",
							},
							"updatedAt": {
								DisplayName: "updatedAt",
								ValueType:   "other",
							},
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
						},
						FieldsMap: map[string]string{},
					},
					"calendars/groups": {
						DisplayName: "Calendars/Groups",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"description": {
								DisplayName: "description",
								ValueType:   "other",
							},
							"slug": {
								DisplayName: "slug",
								ValueType:   "other",
							},
							"isActive": {
								DisplayName: "isActive",
								ValueType:   "other",
							},
							"dateAdded": {
								DisplayName: "dateAdded",
								ValueType:   "other",
							},
							"dateUpdated": {
								DisplayName: "dateUpdated",
								ValueType:   "other",
							},
						},
						FieldsMap: map[string]string{},
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
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"locationId": "iV1BEzddaWWLqU2kXhcN",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
