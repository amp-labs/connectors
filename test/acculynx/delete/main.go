package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// AccuLynx supports the following DELETE operations via the connector:
//   - DELETE /jobs/{jobId}/representatives/ar-owner     (Remove AR Owner)
//   - DELETE /jobs/{jobId}/representatives/sales-owner  (Remove Sales Owner)
//
// AccuLynx exposes no DELETE for top-level /contacts or /jobs — record
// deletion is not part of its V2 API.

// existingJobID is used as the RecordId for both delete operations. Pulled
// from the sandbox; replace if testing against a different AccuLynx account.
const existingJobID = "9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75"

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("=== Test 1: DELETE Job AR Owner ===")

	if err := testDeleteARRepresentative(ctx, conn, existingJobID); err != nil {
		slog.Error("Delete AR owner failed", "error", err)
	} else {
		slog.Info("✅ AR owner removed", "jobId", existingJobID)
	}

	slog.Info("=== Test 2: DELETE Job Sales Owner ===")

	if err := testDeleteSalesRepresentative(ctx, conn, existingJobID); err != nil {
		slog.Error("Delete sales owner failed", "error", err)
	} else {
		slog.Info("✅ Sales owner removed", "jobId", existingJobID)
	}

	slog.Info("All delete tests completed.")
}

func testDeleteARRepresentative(ctx context.Context, conn *acculynx.Connector, jobID string) error {
	params := common.DeleteParams{
		ObjectName: "jobs/representatives/ar-owner",
		RecordId:   jobID,
	}

	slog.Info("Deleting AR owner", "jobId", jobID)

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("delete AR owner: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testDeleteSalesRepresentative(ctx context.Context, conn *acculynx.Connector, jobID string) error {
	params := common.DeleteParams{
		ObjectName: "jobs/representatives/sales-owner",
		RecordId:   jobID,
	}

	slog.Info("Deleting sales owner", "jobId", jobID)

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("delete sales owner: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
