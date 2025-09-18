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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	adTargetingFacetsResponse := testutils.DataFromFile(t, "adTargetingFacets.json")
	dmpEngagementSourceTypesResponse := testutils.DataFromFile(t, "dmpEngagementSourceTypes.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of adTargetingFacets",
			Input: common.ReadParams{ObjectName: "adTargetingFacets", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/adTargetingFacets"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
				},
				Then: mockserver.Response(http.StatusOK, adTargetingFacetsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"entityTypes": []any{
								"COMPANY",
							},
							"facetName":           "followedCompanies",
							"adTargetingFacetUrn": "urn:li:adTargetingFacet:followedCompanies",
							"availableEntityFinders": []any{
								"TYPEAHEAD",
								"SIMILAR_ENTITIES",
							},
						},
					},
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"entityTypes": []any{
								"FIRMOGRAPHIC",
							},
							"facetName":           "revenue",
							"adTargetingFacetUrn": "urn:li:adTargetingFacet:revenue",
							"availableEntityFinders": []any{
								"AD_TARGETING_FACET",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of dmpEngagementSourceTypes",
			Input: common.ReadParams{ObjectName: "dmpEngagementSourceTypes", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/dmpEngagementSourceTypes"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
				},
				Then: mockserver.Response(http.StatusOK, dmpEngagementSourceTypesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"maxLookBack": map[string]any{
								"duration": float64(365),
								"unit":     "DAY",
							},
							"minLookBack": map[string]any{
								"duration": float64(30),
								"unit":     "DAY",
							},
							"engagementSourceTypeDescription": map[string]any{
								"localized": map[string]any{
									"fa_IR": "Conversation",
								},
								"preferredLocale": map[string]any{
									"country":  "US",
									"language": "en",
								},
							},
							"engagementSourceType": "CONVERSATION_ADS",
							"statusMessage": map[string]any{
								"localized": map[string]any{
									"te_IN": "This engagement source type is ready and available to be used in production",
								},
								"preferredLocale": map[string]any{
									"country":  "US",
									"language": "en",
								},
							},
							"engagementSourceTypeStatus": "ACTIVE",
							"defaultLookBack": map[string]any{
								"duration": float64(90),
								"unit":     "DAY",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
