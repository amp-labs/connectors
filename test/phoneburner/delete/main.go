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
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	ownerID, _, err := getOwner(ctx, conn)
	if err != nil {
		utils.Fail("Failed to read members (needed for owner_id)", "error", err)
	}

	// Contacts: create then delete (safe cleanup).
	contactID, err := createContact(ctx, conn, ownerID)
	if err != nil {
		utils.Fail("Failed to create contact for delete", "error", err)
	}
	if err := deleteByID(ctx, conn, "contacts", contactID); err != nil {
		utils.Fail("Failed to delete contact", "error", err, "contact_user_id", contactID)
	}

	// Custom fields: create then delete (safe cleanup).
	customFieldID, err := createCustomField(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create custom field for delete", "error", err)
	}
	if err := deleteByID(ctx, conn, "customfields", customFieldID); err != nil {
		utils.Fail("Failed to delete custom field", "error", err, "custom_field_id", customFieldID)
	}

	// WARNING: phone number delete is very destructive (deletes all numbers).
	// We intentionally do not run it by default.
	if os.Getenv("PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE") == "true" {
		phoneToDelete := os.Getenv("PHONEBURNER_PHONE_NUMBER_TO_DELETE")
		if phoneToDelete == "" {
			utils.Fail("PHONEBURNER_PHONE_NUMBER_TO_DELETE must be set when PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE=true")
		}

		slog.Warn("Deleting phone number (destructive)", "phone_number", phoneToDelete)
		if err := deleteByID(ctx, conn, "phonenumber", phoneToDelete); err != nil {
			utils.Fail("Failed to delete phone number", "error", err, "phone_number", phoneToDelete)
		}
	} else {
		slog.Info("Skipping phonenumber delete (set PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE=true to run)")
	}

	// WARNING: members delete is vendor-level and destructive. Do not run by default.
	if os.Getenv("PHONEBURNER_ALLOW_MEMBER_DELETE") == "true" {
		memberID := os.Getenv("PHONEBURNER_MEMBER_ID_TO_DELETE")
		if memberID == "" {
			utils.Fail("PHONEBURNER_MEMBER_ID_TO_DELETE must be set when PHONEBURNER_ALLOW_MEMBER_DELETE=true")
		}

		slog.Warn("Deleting member (destructive)", "user_id", memberID)
		if err := deleteByID(ctx, conn, "members", memberID); err != nil {
			utils.Fail("Failed to delete member", "error", err, "user_id", memberID)
		}
	} else {
		slog.Info("Skipping members delete (set PHONEBURNER_ALLOW_MEMBER_DELETE=true to run)")
	}

	slog.Info("Delete integration test completed successfully")
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

func createCustomField(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	displayName := fmt.Sprintf("Amp Delete Test Field %s", gofakeit.Word())
	slog.Info("Creating custom field (for delete)", "display_name", displayName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customfields",
		RecordData: map[string]any{
			"display_name":  displayName,
			"type":          1,
			"display_order": 0,
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

func createContact(ctx context.Context, conn *phoneburner.Connector, ownerID string) (string, error) {
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	email := fmt.Sprintf("amp-delete-test-%s@example.com", gofakeit.UUID())
	phone := gofakeit.Numerify("602555####")

	slog.Info("Creating contact (for delete)", "owner_id", ownerID, "email", email)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"owner_id":    ownerID,
			"email":       email,
			"first_name":  firstName,
			"last_name":   lastName,
			"phone":       phone,
			"phone_type":  1,
			"phone_label": "Amp test",
			"notes":       "Created by Ampersand integration delete test",
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
