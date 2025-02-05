package pipedrive

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "zero-records.json")
	success := testutils.DataFromFile(t, "currencies.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "A success API Response",
			Input: []string{"currencies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, success),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"currencies": {
						DisplayName: "Currencies",
						Fields: map[string]common.FieldMetadata{
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
						FieldsMap: map[string]string{
							"active_flag":    "active_flag",
							"code":           "code",
							"decimal_points": "decimal_points",
							"id":             "id",
							"is_custom_flag": "is_custom_flag",
							"name":           "name",
							"symbol":         "symbol",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Zero records returned from server fallback to static file",
			Input: []string{"activities", "leadLabels"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities": {
						DisplayName: "Activities",
						Fields: map[string]common.FieldMetadata{
							"busy_flag": {
								DisplayName:  "busy_flag",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"deal_title": {
								DisplayName:  "deal_title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
						FieldsMap: map[string]string{
							"active_flag":         "active_flag",
							"add_time":            "add_time",
							"assigned_to_user_id": "assigned_to_user_id",
							"attendees":           "attendees",
							"busy_flag":           "busy_flag",
						},
					},
					"leadLabels": {
						DisplayName: "Lead Labels",
						Fields: map[string]common.FieldMetadata{
							"color": {
								DisplayName:  "color",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{
										Value:        "green",
										DisplayValue: "green",
									}, {
										Value:        "blue",
										DisplayValue: "blue",
									}, {
										Value:        "red",
										DisplayValue: "red",
									}, {
										Value:        "yellow",
										DisplayValue: "yellow",
									}, {
										Value:        "purple",
										DisplayValue: "purple",
									}, {
										Value:        "gray",
										DisplayValue: "gray",
									},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(WithAuthenticatedClient(http.DefaultClient))
	if err != nil {
		return nil, err
	}

	connector.setBaseURL(serverURL)

	return connector, nil
}
