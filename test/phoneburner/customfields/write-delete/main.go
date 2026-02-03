package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/phoneburner"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	slog.Info("=== customfields (create -> update -> delete) ===")

	customFieldID, err := createCustomField(ctx, conn)
	if err != nil {
		slog.Error("Failed to create custom field", "error", err)
		return 1
	}
	defer func() {
		if customFieldID != "" {
			_ = cleanupDelete(ctx, conn, "customfields", customFieldID)
		}
	}()

	if err := updateCustomField(ctx, conn, customFieldID); err != nil {
		slog.Error("Failed to update custom field", "error", err, "custom_field_id", customFieldID)
		return 1
	}

	if err := deleteByID(ctx, conn, "customfields", customFieldID); err != nil {
		slog.Error("Failed to delete custom field", "error", err, "custom_field_id", customFieldID)
		return 1
	}
	customFieldID = ""

	slog.Info("PhoneBurner customfields write-delete test completed successfully")
	return 0
}

func createCustomField(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	displayName := fmt.Sprintf("Amp WD Field %s", gofakeit.Word())
	slog.Info("Creating custom field", "display_name", displayName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customfields",
		RecordData: map[string]any{
			"display_name": displayName,
			"type":         1,
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("custom field create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateCustomField(ctx context.Context, conn *phoneburner.Connector, customFieldID string) error {
	displayName := fmt.Sprintf("Amp WD Field Updated %s", gofakeit.Word())
	slog.Info("Updating custom field", "custom_field_id", customFieldID, "display_name", displayName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customfields",
		RecordId:   customFieldID,
		RecordData: map[string]any{
			"display_name": displayName,
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func deleteByID(ctx context.Context, conn *phoneburner.Connector, objectName, recordID string) error {
	slog.Info("Deleting record", "object", objectName, "id", recordID)
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   recordID,
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	if !res.Success {
		return fmt.Errorf("delete reported Success=false")
	}
	return nil
}

func cleanupDelete(ctx context.Context, conn *phoneburner.Connector, objectName, recordID string) error {
	if recordID == "" {
		return nil
	}
	if err := deleteByID(ctx, conn, objectName, recordID); err != nil {
		slog.Warn("Cleanup delete failed", "object", objectName, "id", recordID, "error", err)
		return err
	}
	return nil
}

