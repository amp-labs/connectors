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

	slog.Info("=== folders (create -> update -> delete) ===")

	folderID, err := createFolder(ctx, conn)
	if err != nil {
		slog.Error("Failed to create folder", "error", err)
		return 1
	}
	defer func() {
		if folderID != "" {
			_ = cleanupDelete(ctx, conn, "folders", folderID)
		}
	}()

	if err := updateFolder(ctx, conn, folderID); err != nil {
		slog.Error("Failed to update folder", "error", err, "folder_id", folderID)
		return 1
	}

	if err := deleteByID(ctx, conn, "folders", folderID); err != nil {
		slog.Error("Failed to delete folder", "error", err, "folder_id", folderID)
		return 1
	}
	folderID = ""

	slog.Info("PhoneBurner folders write-delete test completed successfully")
	return 0
}

func createFolder(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	name := fmt.Sprintf("Amp WD Folder %s", gofakeit.Word())
	slog.Info("Creating folder", "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "folders",
		RecordData: map[string]any{
			"name":        name,
			"description": "Created by Ampersand write-delete integration test",
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("folder create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateFolder(ctx context.Context, conn *phoneburner.Connector, folderID string) error {
	name := fmt.Sprintf("Amp WD Folder Updated %s", gofakeit.Word())
	slog.Info("Updating folder", "folder_id", folderID, "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "folders",
		RecordId:   folderID,
		RecordData: map[string]any{
			"name":        name,
			"description": "Updated by Ampersand write-delete integration test",
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
