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

// Creates and updates one amp-test contact as the user in the creds file "metadata.userId".
// The contact is left in place, run the delete test to exercise deletion.
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
	suffix := time.Now().UTC().Format("20060102150405")

	contactID, err := testCreateContact(ctx, conn, suffix)
	if err != nil {
		return err
	}

	return testUpdateContact(ctx, conn, contactID, suffix)
}

func testCreateContact(ctx context.Context, conn *jump.Connector, suffix string) (string, error) {
	slog.Info("Creating contact...")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"firstName": "amp-test",
			"lastName":  "Write " + suffix,
			"email":     fmt.Sprintf("amp-test-write-%s@example.com", suffix),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create contact: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

// updateContact only changes the fields provided, firstName is left unchanged.
func testUpdateContact(ctx context.Context, conn *jump.Connector, contactID, suffix string) error {
	slog.Info("Updating contact...", "id", contactID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: map[string]any{
			"lastName": "Write " + suffix + " Updated",
			"email":    fmt.Sprintf("amp-test-write-%s-updated@example.com", suffix),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
