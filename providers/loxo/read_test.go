package loxo

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

	peopleResponse := testutils.DataFromFile(t, "people.json")
	currenciesResponse := testutils.DataFromFile(t, "currencies.json")
	countriesResponse := testutils.DataFromFile(t, "countries.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of people",
			Input: common.ReadParams{ObjectName: "people", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/integration-user-loxo-withampersand-com/people"),
				Then:  mockserver.Response(http.StatusOK, peopleResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                           float64(1388693),
							"name":                         "George",
							"profile_picture_thumb_url":    "/profile_pictures/thumb/missing.png",
							"profile_picture_original_url": "/profile_pictures/original/missing.png",
							"person_types": []any{
								map[string]any{
									"id":   float64(2157),
									"name": "Candidate",
								},
							},
							"emails": []any{
								map[string]any{
									"id":            float64(922436),
									"value":         "george123@gmail.com",
									"email_type_id": float64(1455),
								},
							},
							"linkedin_url": "https://www.linkedin.com/search/results/people/?keywords=George",
							"source_type": map[string]any{
								"id":   float64(38518),
								"name": "API",
							},
						},
					},
				},
				NextPage: testroutines.URLTestServer +
					"/integration-user-loxo-withampersand-com/people?" +
					"per_page=100&scroll_id=5B313735363231393831313237352C302E302C313338383639335D",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of currencies",
			Input: common.ReadParams{ObjectName: "currencies", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/integration-user-loxo-withampersand-com/currencies"),
				Then:  mockserver.Response(http.StatusOK, currenciesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":        float64(1),
							"code":      "USD",
							"name":      "United States Dollar",
							"symbol":    "$",
							"precision": float64(2),
							"default":   true,
							"position":  float64(1),
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of countries",
			Input: common.ReadParams{ObjectName: "countries", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/integration-user-loxo-withampersand-com/countries"),
				Then:  mockserver.Response(http.StatusOK, countriesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":        float64(3),
							"name":      "Afghanistan",
							"code":      "AF",
							"long_code": "AFG",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/integration-user-loxo-withampersand-com/countries?" +
					"per_page=100&page=101",
				Done: false,
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
