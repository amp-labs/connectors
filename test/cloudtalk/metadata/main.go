package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/cloudtalk"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := cloudtalk.GetCloudTalkConnector(ctx)

	// Test metadata validation against actual API responses.
	// We check 'agents' and 'groups' as these objects are standard and likely to have data in test accounts.
	slog.Info("Testing metadata validation for 'agents'")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "agents", nil)
	slog.Info("✅ Agents metadata validated against API response")

	slog.Info("Testing metadata validation for 'groups'")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "groups", nil)
	slog.Info("✅ Groups metadata validated against API response")

	// Additional validation: Check all objects have metadata defined
	slog.Info("Verifying metadata exists for all objects")

	allObjects := []string{"calls", "agents", "contacts", "groups", "numbers", "tags"}

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
