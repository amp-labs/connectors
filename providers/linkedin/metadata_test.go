package linkedin

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

func TestAdsListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	adTargetingFacetsResponse := testutils.DataFromFile(t, "adTargetingFacets.json")
	dmpEngagementSourceTypesResponse := testutils.DataFromFile(t, "dmpEngagementSourceTypes.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"adTargetingFacets", "dmpEngagementSourceTypes"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/rest/adTargetingFacets"),
						mockcond.Header(http.Header{
							"LinkedIn-Version":          []string{"202504"},
							"X-Restli-Protocol-Version": []string{"2.0.0"},
						}),
					},
					Then: mockserver.Response(http.StatusOK, adTargetingFacetsResponse),
				}, {
					If: mockcond.And{
						mockcond.Path("/rest/dmpEngagementSourceTypes"),
						mockcond.Header(http.Header{
							"LinkedIn-Version":          []string{"202504"},
							"X-Restli-Protocol-Version": []string{"2.0.0"},
						}),
					},
					Then: mockserver.Response(http.StatusOK, dmpEngagementSourceTypesResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"adTargetingFacets": {
						DisplayName: "AdTargetingFacets",
						Fields: map[string]common.FieldMetadata{
							"adTargetingFacetUrn": {
								DisplayName: "adTargetingFacetUrn",
								ValueType:   "other",
							},
							"availableEntityFinders": {
								DisplayName: "availableEntityFinders",
								ValueType:   "other",
							},
							"entityTypes": {
								DisplayName: "entityTypes",
								ValueType:   "other",
							},
							"facetName": {
								DisplayName: "facetName",
								ValueType:   "other",
							},
						},
						FieldsMap: map[string]string{},
					},
					"dmpEngagementSourceTypes": {
						DisplayName: "DmpEngagementSourceTypes",
						Fields: map[string]common.FieldMetadata{
							"defaultLookBack": {
								DisplayName: "defaultLookBack",
								ValueType:   "other",
							},
							"engagementSourceType": {
								DisplayName: "engagementSourceType",
								ValueType:   "other",
							},
							"engagementSourceTypeDescription": {
								DisplayName: "engagementSourceTypeDescription",
								ValueType:   "other",
							},
							"engagementSourceTypeStatus": {
								DisplayName: "engagementSourceTypeStatus",
								ValueType:   "other",
							},
							"maxLookBack": {
								DisplayName: "maxLookBack",
								ValueType:   "other",
							},
							"minLookBack": {
								DisplayName: "minLookBack",
								ValueType:   "other",
							},
							"statusMessage": {
								DisplayName: "statusMessage",
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
				return constructTestAdsConnector(tt.Server.URL)
			})
		})
	}
}
