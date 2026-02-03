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

	slog.Info("=== contacts (create -> update -> delete) ===")

	contactID, err := createContact(ctx, conn, ownerID)
	if err != nil {
		slog.Error("Failed to create contact", "error", err)
		return 1
	}
	defer func() {
		if contactID != "" {
			_ = cleanupDelete(ctx, conn, "contacts", contactID)
		}
	}()

	if err := updateContact(ctx, conn, contactID); err != nil {
		slog.Error("Failed to update contact", "error", err, "contact_user_id", contactID)
		return 1
	}

	if err := deleteByID(ctx, conn, "contacts", contactID); err != nil {
		slog.Error("Failed to delete contact", "error", err, "contact_user_id", contactID)
		return 1
	}
	contactID = ""

	slog.Info("PhoneBurner contacts write-delete test completed successfully")
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

