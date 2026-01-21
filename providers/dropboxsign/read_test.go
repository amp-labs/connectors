package dropboxsign

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	bulkJobsResponse := testutils.DataFromFile(t, "read-bulk_send_job.json")
	faxResponse := testutils.DataFromFile(t, "read-fax.json")
	bulkJobsFirstPage := testutils.DataFromFile(t, "read-bulk_send_job-first-page.json")
	bulkJobsLastPage := testutils.DataFromFile(t, "read-bulk_send_job-last-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "bulk_send_job"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read of bulk_send_job with chosen fields",
			Input: common.ReadParams{ObjectName: "bulk_send_job", Fields: connectors.Fields("bulk_send_job_id", "total", "is_creator", "created_at")}, //nolint:lll
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/bulk_send_job/list"),
				Then:  mockserver.Response(http.StatusOK, bulkJobsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"bulk_send_job_id": "fef03f144d9384737a98ff2ca6c1fd9d7bc2239a",
							"total":            float64(250),
							"is_creator":       false,
							"created_at":       float64(1532740871),
						},
						Raw: map[string]any{
							"bulk_send_job_id": "fef03f144d9384737a98ff2ca6c1fd9d7bc2239a",
							"total":            float64(250),
							"is_creator":       false,
							"created_at":       float64(1532740871),
						},
					},
					{
						Fields: map[string]any{
							"bulk_send_job_id": "6e683bc0369ba3d5b6f43c2c22a8031dbf6bd174",
							"total":            float64(1),
							"is_creator":       true,
							"created_at":       float64(1532640962),
						},
						Raw: map[string]any{
							"bulk_send_job_id": "6e683bc0369ba3d5b6f43c2c22a8031dbf6bd174",
							"total":            float64(1),
							"is_creator":       true,
							"created_at":       float64(1532640962),
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of fax with chosen fields",
			Input: common.ReadParams{
				ObjectName: "fax",
				Fields:     connectors.Fields("fax_id", "title", "subject", "sender"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/fax/list"),
				Then:  mockserver.Response(http.StatusOK, faxResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"fax_id":  "c2e9691c85d9d6fa6ae773842e3680b2b8650f1d",
						"title":   "example title",
						"subject": "example subject",
						"sender":  "me@dropboxsign.com",
					},
					Raw: map[string]any{
						"fax_id":         "c2e9691c85d9d6fa6ae773842e3680b2b8650f1d",
						"title":          "example title",
						"original_title": "example original title",
						"subject":        "example subject",
						"message":        "example message",
						"metadata":       []any{},
						"created_at":     float64(1726774555),
						"sender":         "me@dropboxsign.com",
						"transmissions": []any{
							map[string]any{
								"recipient":   "recipient@dropboxsign.com",
								"sender":      "me@dropboxsign.com",
								"sent_at":     float64(1723231831),
								"status_code": "success",
							},
						},
						"files_url": "https://api.hellosign.com/v3/fax/files/2b388914e3ae3b738bd4e2ee2850c677e6dc53d2",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read first page with pagination",
			Input: common.ReadParams{
				ObjectName: "bulk_send_job",
				Fields:     connectors.Fields("bulk_send_job_id", "total"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/bulk_send_job/list"),
				Then:  mockserver.Response(http.StatusOK, bulkJobsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read last page with no more pages",
			Input: common.ReadParams{
				ObjectName: "bulk_send_job",
				Fields:     connectors.Fields("bulk_send_job_id", "total"),
				NextPage:   "2",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/v3/bulk_send_job/list"),
						mockcond.QueryParam("page", "2"),
					},
					Then: mockserver.Response(http.StatusOK, bulkJobsLastPage),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
