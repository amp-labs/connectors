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

	ownerID, ownerFirstName, err := getOwner(ctx, conn)
	if err != nil {
		utils.Fail("Failed to read members (needed for owner_id)", "error", err)
	}
	slog.Info("Using owner/member", "user_id", ownerID, "first_name", ownerFirstName)

	tagID, err := createTag(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create tag", "error", err)
	}
	// Tag update appears to be disallowed on some accounts (Allow: OPTIONS, DELETE).
	// Skip by default; enable explicitly when needed.
	if os.Getenv("AMP_PB_TEST_TAG_UPDATE") == "1" {
		if err := updateTag(ctx, conn, tagID); err != nil {
			utils.Fail("Failed to update tag", "error", err, "tag_id", tagID)
		}
	} else {
		slog.Warn("Skipping tag update; set AMP_PB_TEST_TAG_UPDATE=1 to enable", "tag_id", tagID)
	}

	folderID, err := createFolder(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create folder", "error", err)
	}
	if err := updateFolder(ctx, conn, folderID); err != nil {
		utils.Fail("Failed to update folder", "error", err, "folder_id", folderID)
	}

	customFieldID, err := createCustomField(ctx, conn)
	if err != nil {
		utils.Fail("Failed to create custom field", "error", err)
	}
	if err := updateCustomField(ctx, conn, customFieldID); err != nil {
		utils.Fail("Failed to update custom field", "error", err, "custom_field_id", customFieldID)
	}

	contactID, err := createContact(ctx, conn, ownerID)
	if err != nil {
		utils.Fail("Failed to create contact", "error", err)
	}
	if err := updateContact(ctx, conn, contactID); err != nil {
		utils.Fail("Failed to update contact", "error", err, "contact_user_id", contactID)
	}

	// Members update can be sensitive. We only do a no-op update (same value)
	// so we exercise PUT /rest/1/members/{user_id} without changing state.
	if err := updateMemberNoop(ctx, conn, ownerID, ownerFirstName); err != nil {
		utils.Fail("Failed to update member (no-op)", "error", err, "user_id", ownerID)
	}

	slog.Info("Write integration test completed successfully")
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

func createTag(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	title := fmt.Sprintf("amp-write-test-tag-%s", gofakeit.UUID())
	slog.Info("Creating tag", "title", title)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"title": title,
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("tag create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateTag(ctx context.Context, conn *phoneburner.Connector, tagID string) error {
	title := fmt.Sprintf("amp-write-test-tag-updated-%s", gofakeit.UUID())
	slog.Info("Updating tag", "tag_id", tagID, "title", title)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordId:   tagID,
		RecordData: map[string]any{
			"title": title,
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func createFolder(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	name := fmt.Sprintf("Amp Write Test Folder %s", gofakeit.Word())
	slog.Info("Creating folder", "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "folders",
		RecordData: map[string]any{
			// PhoneBurner expects JSON body for folders.
			"name":        name,
			"description": "Created by Ampersand integration test",
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
	name := fmt.Sprintf("Amp Write Test Folder Updated %s", gofakeit.Word())
	slog.Info("Updating folder", "folder_id", folderID, "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "folders",
		RecordId:   folderID,
		RecordData: map[string]any{
			"name":        name,
			"description": "Updated by Ampersand integration test",
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func createCustomField(ctx context.Context, conn *phoneburner.Connector) (string, error) {
	displayName := fmt.Sprintf("Amp Write Test Field %s", gofakeit.Word())
	slog.Info("Creating custom field", "display_name", displayName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customfields",
		RecordData: map[string]any{
			"display_name":  displayName,
			"type":          1, // Text field
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

func updateCustomField(ctx context.Context, conn *phoneburner.Connector, customFieldID string) error {
	displayName := fmt.Sprintf("Amp Write Test Field Updated %s", gofakeit.Word())
	slog.Info("Updating custom field", "custom_field_id", customFieldID, "display_name", displayName)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customfields",
		RecordId:   customFieldID,
		RecordData: map[string]any{
			"display_name":  displayName,
			"type":          1,
			"display_order": 0,
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
	email := fmt.Sprintf("amp-write-test-%s@example.com", gofakeit.UUID())
	phone := gofakeit.Numerify("602555####")

	slog.Info("Creating contact", "owner_id", ownerID, "email", email)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			// PhoneBurner expects form body for contacts.
			"owner_id":    ownerID,
			"email":       email,
			"first_name":  firstName,
			"last_name":   lastName,
			"phone":       phone,
			"phone_type":  1,
			"phone_label": "Amp test",
			"notes":       "Created by Ampersand integration test",
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
			"notes":      "Updated by Ampersand integration test",
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func updateMemberNoop(ctx context.Context, conn *phoneburner.Connector, userID string, firstName string) error {
	if firstName == "" {
		// If the API doesn't return first_name (sometimes empty), pick a conservative no-op.
		firstName = " "
	}
	slog.Info("Updating member (no-op)", "user_id", userID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "members",
		RecordId:   userID,
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
