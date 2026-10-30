package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/jump"
	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
)

// Creates one amp-test contact as the user in the creds file "metadata.userId" and deletes it.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetJumpConnector(ctx)

	if err := run(ctx, conn); err != nil {
		utils.Fail("error", "error", err)
	}
}

func run(ctx context.Context, conn *jump.Connector) error {
	contactID, err := createTestContact(ctx, conn)
	if err != nil {
		return err
	}

	return testDeleteContact(ctx, conn, contactID)
}

func createTestContact(ctx context.Context, conn *jump.Connector) (string, error) {
	suffix := time.Now().UTC().Format("20060102150405")

	slog.Info("Creating test contact for delete...")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"firstName": "amp-test",
			"lastName":  "Delete " + suffix,
			"email":     fmt.Sprintf("amp-test-delete-%s@example.com", suffix),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create contact: %w", err)
	}

	slog.Info("Created contact", "id", res.RecordId)

	return res.RecordId, nil
}

func testDeleteContact(ctx context.Context, conn *jump.Connector, contactID string) error {
	slog.Info("Deleting contact...", "id", contactID)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
