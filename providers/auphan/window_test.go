package auphan

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// reportBody is the reports/execute body for payments with one "payment" column.
func reportBody(filters string) string {
	return `{"store_id": 73921, "params": {"type": "payments", "columns": ["payment"], "filters": ` + filters + `}}`
}

func TestReadReportWindows(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseEmpty := testutils.DataFromFile(t, "payments-empty.json")
	marchMid := time.Date(2000, 3, 15, 0, 0, 0, 0, time.UTC)
	emptyPage := func(next string) *common.ReadResult {
		return &common.ReadResult{
			Rows: 0, Data: []common.ReadResultRow{},
			NextPage: common.NextPageToken(next), Done: next == "",
		}
	}

	tests := []struct {
		testconn.TestCaseRead

		metadata map[string]string
	}{
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Without Since the read starts in 2000, one 30-day window at a time",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"), Until: marchMid,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(reportBody(`{"selection": "close_date",
					"date_from": "2000-01-01 00:00:00", "date_to": "2000-01-30 23:59:59"}`)),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: emptyPage("2000-01-31T00:00:00Z/2000-03-15T00:00:00Z"),
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "The last window ends at Until and is done",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"),
				NextPage: "2000-03-01T00:00:00Z/2000-03-15T00:00:00Z",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(reportBody(`{"selection": "close_date",
					"date_from": "2000-03-01 00:00:00", "date_to": "2000-03-15 00:00:00"}`)),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: emptyPage(""),
		}},
		{
			TestCaseRead: testconn.TestCaseRead{
				Name: "UTC times are sent as the store's local day",
				Input: common.ReadParams{
					ObjectName: "payments", Fields: connectors.Fields("payment"),
					Since: time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC),
					Until: time.Date(2026, 11, 6, 7, 59, 59, 0, time.UTC),
				},
				Server: mockserver.Conditional{
					Setup: mockserver.ContentJSON(),
					If: mockcond.Body(reportBody(`{"selection": "close_date",
						"date_from": "2026-11-05 00:00:00", "date_to": "2026-11-05 23:59:59"}`)),
					Then: mockserver.Response(http.StatusOK, responseEmpty),
				}.Server(),
				Expected: emptyPage(""),
			},
			metadata: map[string]string{"timezone": "America/Los_Angeles"},
		},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Report filters are merged into the request",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"),
				Since: marchMid, Until: marchMid.Add(time.Hour),
				RawFilter: &common.RawFilter{
					Type:   RawFilterTypeReportFilters,
					Filter: map[string]any{"selection": "open_date", "status": []any{2}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(reportBody(`{"selection": "open_date", "status": [2],
					"date_from": "2000-03-15 00:00:00", "date_to": "2000-03-15 01:00:00"}`)),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: emptyPage(""),
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Report filters cannot set the date range",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"), Since: marchMid,
				RawFilter: &common.RawFilter{
					Type: RawFilterTypeReportFilters, Filter: map[string]any{"date_from": "2000-01-01 00:00:00"},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrInvalidRawFilter, errFilterDateRange},
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Report filters are rejected for stores",
			Input: common.ReadParams{
				ObjectName: "stores", Fields: connectors.Fields("store_id"),
				RawFilter: &common.RawFilter{Type: RawFilterTypeReportFilters, Filter: map[string]any{}},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrInvalidRawFilter, errFilterOnlyReports},
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Since after Until is rejected",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"),
				Since: marchMid, Until: marchMid.Add(-time.Hour),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrSinceUntilChronOrder},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnectorWithMetadata(tt.Server.URL, tt.metadata)
			})
		})
	}
}

func TestInvalidTimezone(t *testing.T) {
	t.Parallel()

	_, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "73921",
		Metadata:            map[string]string{"timezone": "Mars/Olympus_Mons"},
	})
	if !errors.Is(err, errInvalidTimezone) {
		t.Fatalf("expected errInvalidTimezone, got %v", err)
	}
}
