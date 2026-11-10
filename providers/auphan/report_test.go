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

func TestReadReportFilters(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseEmpty := testutils.DataFromFile(t, "payments-empty.json")
	marchMid := time.Date(2000, 3, 15, 0, 0, 0, 0, time.UTC)
	done := &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true}

	tests := []struct {
		testconn.TestCaseRead

		metadata map[string]string
	}{
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Since and Until are sent as the date range, in one call",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"),
				Since: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Until: marchMid,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(reportBody(`{"selection": "close_date",
					"date_from": "2000-01-01 00:00:00", "date_to": "2000-03-15 00:00:00"}`)),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Without Until no date_to is sent, so Auphan reads through today",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"), Since: marchMid,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Body(reportBody(`{"selection": "close_date", "date_from": "2000-03-15 00:00:00"}`)),
				Then:  mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "Without Since the read starts at 2000-01-01, so the whole history is read",
			Input: common.ReadParams{
				ObjectName: "payments", Fields: connectors.Fields("payment"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Body(reportBody(`{"selection": "close_date", "date_from": "2000-01-01 00:00:00"}`)),
				Then:  mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "A report's required filters are sent, and it gets no selection when it has none",
			Input: common.ReadParams{
				ObjectName: "attendance", Fields: connectors.Fields("employee"), Since: marchMid, Until: marchMid,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"store_id": 73921, "params": {"type": "attendance", "columns": ["employee"],
					"filters": {"status": [4], "date_from": "2000-03-15 00:00:00", "date_to": "2000-03-15 00:00:00"}}}`),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "A report on dates gets date-only bounds",
			Input: common.ReadParams{
				ObjectName: "salerules", Fields: connectors.Fields("salerule_name"),
				Since: marchMid, Until: marchMid.AddDate(0, 0, 1),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"store_id": 73921, "params": {"type": "salerules", "columns": ["salerule_name"],
					"filters": {"date_from": "2000-03-15", "date_to": "2000-03-16"}}}`),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{TestCaseRead: testconn.TestCaseRead{
			Name: "A report without a date range gets no dates, even with Since",
			Input: common.ReadParams{
				ObjectName: "usergroups", Fields: connectors.Fields("name"), Since: marchMid,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{"store_id": 73921, "params": {"type": "usergroups", "columns": ["name"],
					"filters": {}}}`),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: done,
		}},
		{
			TestCaseRead: testconn.TestCaseRead{
				Name: "The 2000-01-01 default is not shifted by the store timezone",
				Input: common.ReadParams{
					ObjectName: "salerules", Fields: connectors.Fields("salerule_name"), Until: marchMid,
				},
				Server: mockserver.Conditional{
					Setup: mockserver.ContentJSON(),
					If: mockcond.Body(`{"store_id": 73921, "params": {"type": "salerules", "columns": ["salerule_name"],
						"filters": {"date_from": "2000-01-01", "date_to": "2000-03-14"}}}`),
					Then: mockserver.Response(http.StatusOK, responseEmpty),
				}.Server(),
				Expected: done,
			},
			metadata: map[string]string{"timezone": "America/Los_Angeles"},
		},
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
				Expected: done,
			},
			metadata: map[string]string{"timezone": "America/Los_Angeles"},
		},
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
