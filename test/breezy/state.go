// Package breezy provides live-test helpers for the Breezy HR connector.
//
// # Position state helpers (test-only)
//
// PublishPosition and SetPositionState are integration-test utilities. They are not
// part of any connector interface (Read/Write/Delete) and are not exposed to customers.
//
// Why they exist:
//   - Breezy creates new positions in draft state.
//   - The positions list endpoint defaults to state=published, so draft records are
//     invisible to the connector Read used by ValidateCreateUpdateDelete.
//   - PostCreate in the write-delete live test calls PublishPosition so the created
//     record appears in the default list before read/update validation runs.
//
// The connector Delete operation uses the same state endpoint with state=archived
// (soft delete). PublishPosition is the complementary test setup step (draft → published).
//
// API reference: https://developer.breezy.hr/reference/company-position-state-update
package breezy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/breezy"
)

// SetPositionState updates a position lifecycle state via PUT …/position/{id}/state.
// https://developer.breezy.hr/reference/company-position-state-update
func SetPositionState(ctx context.Context, conn *breezy.Connector, positionID, state string) error {
	u, err := urlbuilder.New(
		conn.ProviderInfo().BaseURL,
		"v3",
		"company",
		conn.CompanyID,
		"position",
		positionID,
		"state",
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"state": state})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := conn.HTTPClient().Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to set position state %q: HTTP %d", state, resp.StatusCode)
	}

	return nil
}

// PublishPosition moves a draft position to published so it appears in the default positions list.
func PublishPosition(ctx context.Context, conn *breezy.Connector, positionID string) error {
	return SetPositionState(ctx, conn, positionID, "published")
}
