package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
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

	ownerID, ownerFirstName, err := getOwner(ctx, conn)
	if err != nil {
		slog.Error("Failed to read members (needed for owner_id)", "error", err)
		return 1
	}
	slog.Info("Using owner/member", "user_id", ownerID, "first_name", ownerFirstName)

	var (
		folderID      string
		customFieldID string
		contactID     string
	)
	defer func() {
		if contactID != "" {
			_ = cleanupDelete(ctx, conn, "contacts", contactID)
		}
		if customFieldID != "" {
			_ = cleanupDelete(ctx, conn, "customfields", customFieldID)
		}
		if folderID != "" {
			_ = cleanupDelete(ctx, conn, "folders", folderID)
		}
	}()

	// ------------------------------------------------------------
	// Step 1: Folder create -> update -> delete
	// ------------------------------------------------------------
	slog.Info("=== Step 1: folders (create -> update -> delete) ===")
	folderID, err = createFolder(ctx, conn)
	if err != nil {
		slog.Error("Failed to create folder", "error", err)
		return 1
	}
	if err := updateFolder(ctx, conn, folderID); err != nil {
		slog.Error("Failed to update folder", "error", err, "folder_id", folderID)
		return 1
	}
	if err := deleteByID(ctx, conn, "folders", folderID); err != nil {
		slog.Error("Failed to delete folder", "error", err, "folder_id", folderID)
		return 1
	}
	folderID = ""

	// ------------------------------------------------------------
	// Step 2: Custom fields create -> update -> delete
	// ------------------------------------------------------------
	slog.Info("=== Step 2: customfields (create -> update -> delete) ===")
	customFieldID, err = createCustomField(ctx, conn)
	if err != nil {
		slog.Error("Failed to create custom field", "error", err)
		return 1
	}
	if err := updateCustomField(ctx, conn, customFieldID); err != nil {
		slog.Error("Failed to update custom field", "error", err, "custom_field_id", customFieldID)
		return 1
	}
	if err := deleteByID(ctx, conn, "customfields", customFieldID); err != nil {
		slog.Error("Failed to delete custom field", "error", err, "custom_field_id", customFieldID)
		return 1
	}
	customFieldID = ""

	// ------------------------------------------------------------
	// Step 3: Contacts create -> update -> delete
	// ------------------------------------------------------------
	slog.Info("=== Step 3: contacts (create -> update -> delete) ===")
	contactID, err = createContact(ctx, conn, ownerID)
	if err != nil {
		slog.Error("Failed to create contact", "error", err)
		return 1
	}
	if err := updateContact(ctx, conn, contactID); err != nil {
		slog.Error("Failed to update contact", "error", err, "contact_user_id", contactID)
		return 1
	}
	if err := deleteByID(ctx, conn, "contacts", contactID); err != nil {
		slog.Error("Failed to delete contact", "error", err, "contact_user_id", contactID)
		return 1
	}
	contactID = ""

	slog.Info("PhoneBurner write-delete integration test completed successfully")
	return 0
}

func getOwner(ctx context.Context, conn *phoneburner.Connector) (userID string, firstName string, err error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "members",
		Fields:     connectors.Fields("user_id", "first_name"),
		PageSize:   1,
	})
	if err != nil {
		return "", "", err
	}
	if len(res.Data) == 0 {
		return "", "", fmt.Errorf("no members returned")
	}

	id, _ := res.Data[0].Raw["user_id"].(string)
	if id == "" {
		return "", "", fmt.Errorf("members response missing user_id")
	}
	fn, _ := res.Data[0].Raw["first_name"].(string)
	return id, fn, nil
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

func createContact(ctx context.Context, conn *phoneburner.Connector, ownerID string) (string, error) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	email := fmt.Sprintf("amp-wd-%s@example.com", gofakeit.UUID())
	phone := gofakeit.Numerify("602555####")

	slog.Info("Creating contact", "owner_id", ownerID, "email", email)
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"owner_id":    ownerID,
			"email":       email,
			"first_name":  firstName,
			"last_name":   lastName,
			"phone":       phone,
			"phone_type":  1,
			"phone_label": "Amp wd test",
			"notes":       "Created by Ampersand write-delete integration test",
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)
	if res.RecordId == "" {
		return "", fmt.Errorf("contact create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateContact(ctx context.Context, conn *phoneburner.Connector, contactID string) error {
	firstName := gofakeit.FirstName()
	slog.Info("Updating contact", "contact_user_id", contactID, "first_name", firstName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: map[string]any{
			"first_name": firstName,
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
