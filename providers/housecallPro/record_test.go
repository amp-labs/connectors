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

func TestGetRecordsByIds_invoice_usesAPIPath(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "id": "invoice_5f3de0f1d9e1483f9a4e4be0c7a44f0b",
  "status": "paid",
  "invoice_number": "INV-1042"
}`)

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Method(http.MethodGet),
			mockcond.Path("/api/invoices/invoice_5f3de0f1d9e1483f9a4e4be0c7a44f0b"),
		},
		Then: mockserver.Response(http.StatusOK, payload),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := conn.GetRecordsByIds(t.Context(), "invoices",
		[]string{"invoice_5f3de0f1d9e1483f9a4e4be0c7a44f0b"},
		[]string{"id", "status", "invoice_number"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 || rows[0].Id != "invoice_5f3de0f1d9e1483f9a4e4be0c7a44f0b" {
		t.Fatalf("unexpected row: %+v", rows)
	}
}

func TestGetRecordsByIds_unknownObject(t *testing.T) {
	t.Parallel()

	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.GetRecordsByIds(t.Context(), "not_a_connector_object", []string{"x"}, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
