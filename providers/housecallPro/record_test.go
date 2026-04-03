package housecallpro

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

func TestGetRecordsByIds_job(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "id": "job_ac6f3efd11c14a5aa93e9fc0ab5354ab",
  "work_status": "needs scheduling",
  "updated_at": "2026-04-02T15:20:38Z"
}`)

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodGet),
			mockcond.Path("/jobs/job_ac6f3efd11c14a5aa93e9fc0ab5354ab"),
		},
		Then: mockserver.Response(http.StatusOK, payload),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := conn.GetRecordsByIds(t.Context(), "jobs",
		[]string{"job_ac6f3efd11c14a5aa93e9fc0ab5354ab"},
		[]string{"id", "work_status"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 || rows[0].Id != "job_ac6f3efd11c14a5aa93e9fc0ab5354ab" {
		t.Fatalf("unexpected row: %+v", rows)
	}
}

func TestGetRecordsByIds_unsupportedObject(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.GetRecordsByIds(t.Context(), "invoices", []string{"x"}, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
