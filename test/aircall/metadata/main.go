package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/aircall"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := aircall.GetAircallConnector(ctx)

	// Test metadata validation against actual API responses
	// This makes real API calls and validates our schema matches reality

	slog.Info("Testing metadata validation for 'users' (has data)")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "users", nil)
	slog.Info("✅ Users metadata validated against API response")

	// Note: We only test 'users' because:
	// - ValidateMetadataContainsRead requires at least 1 record to exist
	// - The test account has users but may not have calls/contacts/numbers
	// - This validates the pattern works; other objects follow same structure

	// Additional validation: Check all objects have metadata defined
	slog.Info("Verifying metadata exists for all objects")
	allObjects := []string{"calls", "users", "contacts", "numbers", "teams", "tags"}

	for _, objectName := range allObjects {
		metadata, err := conn.ListObjectMetadata(ctx, []string{objectName})
		if err != nil {
			slog.Error("Failed to get metadata", "object", objectName, "error", err)
			os.Exit(1)
		}

		objectMeta := metadata.Result[objectName]
		fieldCount := len(objectMeta.Fields)

		if fieldCount == 0 {
			slog.Error("No fields defined for object", "object", objectName)
			os.Exit(1)
		}

		slog.Info("Metadata defined",
			"object", objectName,
			"displayName", objectMeta.DisplayName,
			"fieldCount", fieldCount)
	}

	slog.Info("✅ All metadata tests passed")
}
