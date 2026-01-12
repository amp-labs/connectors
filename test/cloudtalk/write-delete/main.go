package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/cloudtalk"
	testCloudTalk "github.com/amp-labs/connectors/test/cloudtalk"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testCloudTalk.GetCloudTalkConnector(ctx)

	slog.Info("Testing write/delete for contacts")
	if err := testWriteDeleteContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("Testing write/delete for tags")
	if err := testWriteDeleteTags(ctx, conn); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func testWriteDeleteTags(ctx context.Context, conn *cloudtalk.Connector) error {
	// 1. Create Tag
	slog.Info("Creating tag...")
	createParams := common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"name": fmt.Sprintf("Tag-%s", gofakeit.UUID()),
		},
	}

	createRes, err := conn.Write(ctx, createParams)
	if err != nil {
		return fmt.Errorf("error creating tag: %w", err)
	}

	if !createRes.Success {
		return fmt.Errorf("failed to create tag")
	}

	tagID := createRes.RecordId
	slog.Info("Tag created", "id", tagID, "data", createRes.Data)

	// 2. Update Tag
	slog.Info("Updating tag...", "id", tagID)
	updateParams := common.WriteParams{
		ObjectName: "tags",
		RecordId:   tagID,
		RecordData: map[string]any{
			"name": fmt.Sprintf("Tag-Updated-%s", gofakeit.UUID()),
		},
	}

	updateRes, err := conn.Write(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("error updating tag: %w", err)
	}

	if !updateRes.Success {
		return fmt.Errorf("failed to update tag")
	}

	slog.Info("Tag updated", "id", updateRes.RecordId, "data", updateRes.Data)

	// 3. Delete Tag
	slog.Info("Deleting tag...", "id", tagID)
	deleteParams := common.DeleteParams{
		ObjectName: "tags",
		RecordId:   tagID,
	}

	deleteRes, err := conn.Delete(ctx, deleteParams)
	if err != nil {
		return fmt.Errorf("error deleting tag: %w", err)
	}

	if !deleteRes.Success {
		return fmt.Errorf("failed to delete tag")
	}

	slog.Info("Tag deleted successfully")

	return nil
}

func testWriteDeleteContacts(ctx context.Context, conn *cloudtalk.Connector) error {
	// 1. Create Contact
	slog.Info("Creating contact...")
	createParams := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"name":  gofakeit.Name(),
			"email": gofakeit.Email(),
		},
	}

	createRes, err := conn.Write(ctx, createParams)
	if err != nil {
		return fmt.Errorf("error creating contact: %w", err)
	}

	if !createRes.Success {
		return fmt.Errorf("failed to create contact")
	}

	contactID := createRes.RecordId
	slog.Info("Contact created", "id", contactID, "data", createRes.Data)

	// 2. Update Contact
	slog.Info("Updating contact...", "id", contactID)
	updateParams := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: map[string]any{
			"name": gofakeit.Name(),
		},
	}

	updateRes, err := conn.Write(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("error updating contact: %w", err)
	}

	if !updateRes.Success {
		return fmt.Errorf("failed to update contact")
	}

	slog.Info("Contact updated", "id", updateRes.RecordId, "data", updateRes.Data)

	// 3. Delete Contact
	slog.Info("Deleting contact...", "id", contactID)
	deleteParams := common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
	}

	deleteRes, err := conn.Delete(ctx, deleteParams)
	if err != nil {
		return fmt.Errorf("error deleting contact: %w", err)
	}

	if !deleteRes.Success {
		return fmt.Errorf("failed to delete contact")
	}

	slog.Info("Contact deleted successfully")

	return nil
}
