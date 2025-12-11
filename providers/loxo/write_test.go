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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	companiesFieldResponse := testutils.DataFromFile(t, "write_companies.json")
	peopleResponse := testutils.DataFromFile(t, "write_people.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the companies",
			Input: common.WriteParams{ObjectName: "companies", RecordData: map[string]any{"company[name]": "sample value"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/integration-user-loxo-withampersand-com/companies"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, companiesFieldResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1359599",
				Errors:   nil,
				Data: map[string]any{
					"id":             float64(1359599),
					"name":           "Lever",
					"logo_large_url": "/logos/large/missing.png",
					"logo_thumb_url": "/logos/thumb/missing.png",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the people",
			Input: common.WriteParams{ObjectName: "people", RecordData: map[string]any{"person[name]": "Sample"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/integration-user-loxo-withampersand-com/people"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, peopleResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1659529",
				Errors:   nil,
				Data: map[string]any{
					"id":                           float64(1659529),
					"name":                         "Sam",
					"profile_picture_thumb_url":    "/profile_pictures/thumb/missing.png",
					"profile_picture_original_url": "/profile_pictures/original/missing.png",
					"linkedin_url":                 "https://www.linkedin.com/search/results/people/?keywords=Sam",
					"agency_id":                    float64(39026),
					"person_types": []any{
						map[string]any{
							"id":       float64(2157),
							"key":      "candidate",
							"name":     "Candidate",
							"default":  true,
							"position": float64(1),
						},
					},
					"source_type": map[string]any{
						"id":   float64(38518),
						"name": "API",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Updating the people",
			Input: common.WriteParams{
				ObjectName: "people",
				RecordData: map[string]any{"person[name]": "Sample"},
				RecordId:   "1659529",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/integration-user-loxo-withampersand-com/people/1659529"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, peopleResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1659529",
				Errors:   nil,
				Data: map[string]any{
					"id":                           float64(1659529),
					"name":                         "Sam",
					"profile_picture_thumb_url":    "/profile_pictures/thumb/missing.png",
					"profile_picture_original_url": "/profile_pictures/original/missing.png",
					"linkedin_url":                 "https://www.linkedin.com/search/results/people/?keywords=Sam",
					"agency_id":                    float64(39026),
					"person_types": []any{
						map[string]any{
							"id":       float64(2157),
							"key":      "candidate",
							"name":     "Candidate",
							"default":  true,
							"position": float64(1),
						},
					},
					"source_type": map[string]any{
						"id":   float64(38518),
						"name": "API",
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
