package pipedrive

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "zero-records.json")
	success := testutils.DataFromFile(t, "currencies.json")
	activityFields := testutils.DataFromFile(t, "activityFields.json")

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
			Input: []string{"leadLabels"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
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
		{
			Name:  "Objects using metadata discovery endpoints",
			Input: []string{"activities"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/activityFields"),
				Then:  mockserver.Response(http.StatusOK, activityFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities": {
						DisplayName: "Activities",
						Fields: map[string]common.FieldMetadata{
							"priority": {
								DisplayName:  "Priority",
								ValueType:    "singleSelect",
								ProviderType: "enum",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{
									{
										Value:        "24",
										DisplayValue: "Low",
									}, {
										Value:        "25",
										DisplayValue: "Medium",
									}, {
										Value:        "26",
										DisplayValue: "High",
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
	connector, err := NewConnector(WithAuthenticatedClient(mockutils.NewClient()))
	if err != nil {
		return nil, err
	}

	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
