package closecrm

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "zero-records.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	activityResponse := testutils.DataFromFile(t, "activities.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "united", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				testutils.StringError(string(unsupportedResponse)),
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "opportunity_stages", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully read activities",
			Input: common.ReadParams{
				ObjectName: "activity",
				Fields:     connectors.Fields("user_id", "user_name", "source", "id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v1/activity/"),
				Then:  mockserver.Response(http.StatusOK, activityResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":        "acti_4440QLsync5No96XBuQkLfCngfKGXXixrWtWVUjM6lv",
						"source":    "manual",
						"user_id":   "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
						"user_name": "Josep Karage",
					},
					Raw: map[string]any{
						"_type":             "Meeting",
						"activity_at":       "2025-01-20T10:43:46.561000+00:00",
						"actual_duration":   nil,
						"attached_call_ids": []any{},
						"date_created":      "2025-01-20T10:43:46.561000+00:00",
						"date_updated":      "2025-01-14T10:44:08.137000+00:00",
						"ends_at":           "2025-01-20T11:28:46.561000+00:00",
						"id":                "acti_4440QLsync5No96XBuQkLfCngfKGXXixrWtWVUjM6lv",
						"is_recurring":      false,
						"lead_id":           "lead_G7FYn5pkohGlQzAgddQ9zcLOoYSkP2PIFsflsSOMq71",
						"organization_id":   "orga_Bnvb4C6ur74cnEX825rjjDDusTIIpQcFX2qusZcjGr5",
						"source":            "manual",
						"starts_at":         "2025-01-20T10:43:46.561000+00:00",
						"title":             "Contract Finalization",
						"user_id":           "user_4N6K1GpqhrjELJKSIMos60M78s2x86Qy6jAiUht5tmh",
						"user_name":         "Josep Karage",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// TestReadPagination follows NextPage across three pages.
// Each page must request the next offset, otherwise the same page is returned forever.
// A third page catches offsets built by joining strings ("0"+"100"+"100") instead of adding numbers.
func TestReadPagination(t *testing.T) { // nolint:funlen
	t.Parallel()

	pages := [][]byte{
		testutils.DataFromFile(t, "leads-page-1.json"),
		testutils.DataFromFile(t, "leads-page-2.json"),
		testutils.DataFromFile(t, "leads-page-3.json"),
	}

	pageCase := func(skip string, body []byte) mockserver.Case {
		return mockserver.Case{
			If: mockcond.And{
				mockcond.Path("/api/v1/lead/"),
				mockcond.QueryParam("_skip", skip),
				mockcond.QueryParam("_limit", "100"),
			},
			Then: mockserver.Response(http.StatusOK, body),
		}
	}

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			pageCase("0", pages[0]),
			pageCase("100", pages[1]),
			pageCase("200", pages[2]),
		},
	}.Server()
	defer server.Close()

	connector, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	params := common.ReadParams{
		ObjectName: "lead",
		Fields:     connectors.Fields("id", "display_name"),
	}

	expectedIDs := []string{"lead_page_1", "lead_page_2", "lead_page_3"}

	for pageIndex, expectedID := range expectedIDs {
		result, err := connector.Read(t.Context(), params)
		if err != nil {
			t.Fatalf("page %d: unexpected error: %v", pageIndex+1, err)
		}

		if result.Rows != 1 {
			t.Fatalf("page %d: expected 1 row, got %d", pageIndex+1, result.Rows)
		}

		if id := result.Data[0].Fields["id"]; id != expectedID {
			t.Fatalf("page %d: expected record %q, got %q", pageIndex+1, expectedID, id)
		}

		isLastPage := pageIndex == len(expectedIDs)-1
		if result.Done != isLastPage {
			t.Fatalf("page %d: expected done=%v, got %v", pageIndex+1, isLastPage, result.Done)
		}

		params.NextPage = result.NextPage
	}
}
