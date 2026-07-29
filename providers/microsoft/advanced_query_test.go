package microsoft

import (
	"context"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
)

// TestNeedsAdvancedQuery guards the directory-object set against drift from the
// authoritative Graph list (https://learn.microsoft.com/graph/aad-advanced-queries).
func TestNeedsAdvancedQuery(t *testing.T) {
	t.Parallel()

	directory := []string{
		"users", "groups", "applications", "servicePrincipals", "devices",
		"administrativeUnits", "contacts", "appRoleAssignments", "oauth2PermissionGrants",
		// Delta-query variants filter on createdDateTime too, so they need advanced query.
		"users/microsoft.graph.delta()", "groups/microsoft.graph.delta()",
		"applications/microsoft.graph.delta()",
	}
	for _, obj := range directory {
		if !needsAdvancedQuery(obj) {
			t.Errorf("needsAdvancedQuery(%q) = false, want true", obj)
		}
	}

	// Not directory objects (mail/calendar/drive), or a directory role that is not
	// in the authoritative advanced-query list.
	notDirectory := []string{"messages", "events", "drive/items", "alerts", "directoryRoles"}
	for _, obj := range notDirectory {
		if needsAdvancedQuery(obj) {
			t.Errorf("needsAdvancedQuery(%q) = true, want false", obj)
		}
	}
}

// TestBuildReadRequestAdvancedQuery verifies directory-object incremental reads
// carry the advanced-query parameters Graph requires ($count=true +
// ConsistencyLevel: eventual), and that the header is present on all pages.
func TestBuildReadRequestAdvancedQuery(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("https://graph.microsoft.com")
	if err != nil {
		t.Fatalf("constructTestConnector: %v", err)
	}

	since := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("directory object with Since gets $count + ConsistencyLevel + $filter", func(t *testing.T) {
		t.Parallel()

		req, err := conn.buildReadRequest(context.Background(), common.ReadParams{ObjectName: "groups", Since: since})
		if err != nil {
			t.Fatalf("buildReadRequest: %v", err)
		}

		query := req.URL.Query()
		if got := query.Get("$count"); got != "true" {
			t.Errorf("$count = %q, want true", got)
		}

		if query.Get("$filter") == "" {
			t.Error("$filter should be set for an incremental read")
		}

		if got := req.Header.Get("ConsistencyLevel"); got != "eventual" {
			t.Errorf("ConsistencyLevel = %q, want eventual", got)
		}
	})

	t.Run("directory object without Since keeps the header but omits $count", func(t *testing.T) {
		t.Parallel()

		req, err := conn.buildReadRequest(context.Background(), common.ReadParams{ObjectName: "groups"})
		if err != nil {
			t.Fatalf("buildReadRequest: %v", err)
		}

		if got := req.URL.Query().Get("$count"); got != "" {
			t.Errorf("$count = %q, want empty (no filter)", got)
		}

		// The header must ride every page, including @odata.nextLink follow-ups.
		if got := req.Header.Get("ConsistencyLevel"); got != "eventual" {
			t.Errorf("ConsistencyLevel = %q, want eventual", got)
		}
	})
}
