package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/getresponse"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

// MainFn runs the contacts write-delete E2E: get campaign → create contact → update → delete.
// Contacts must be associated with a campaign (GetResponse API).
func MainFn() int {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Step 1: Get a campaign from the list (required for creating contacts).
	slog.Info("=== Test 1: Get a campaign from the list ===")
	campaignID, err := getCampaignFromList(ctx, conn)
	if err != nil {
		utils.Fail("Failed to get campaign from list", "error", err)
	}

	// Step 2: Create contact.
	slog.Info("=== Test 2: Create contact ===")
	contactID, err := createContact(ctx, conn, campaignID)
	if err != nil {
		utils.Fail("Failed to create contact", "error", err)
	}
	slog.Info("Contact created", "contactId", contactID)

	time.Sleep(2 * time.Second)

	// Step 3: Update contact.
	slog.Info("=== Test 3: Update contact ===")
	if err := updateContact(ctx, conn, contactID); err != nil {
		utils.Fail("Failed to update contact", "error", err)
	}
	slog.Info("Contact updated")

	// Step 4: Delete contact (cleanup).
	slog.Info("=== Test 4: Delete contact ===")
	if err := deleteContact(ctx, conn, contactID); err != nil {
		utils.Fail("Failed to delete contact", "error", err)
	}
	slog.Info("Contact deleted")

	slog.Info("Contacts write-delete tests completed successfully!")
	return 0
}

// getCampaignFromList lists campaigns and returns the first campaign ID.
func getCampaignFromList(ctx context.Context, conn *getresponse.Connector) (string, error) {
	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "isDefault", "createdOn"),
		PageSize:   10,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", fmt.Errorf("reading campaigns: %w", err)
	}

	if res.Rows == 0 || len(res.Data) == 0 {
		return "", fmt.Errorf("no campaigns found in account")
	}

	campaignID, ok := res.Data[0].Raw["campaignId"].(string)
	if !ok {
		return "", fmt.Errorf("campaignId not found in response")
	}

	name, _ := res.Data[0].Raw["name"].(string)
	slog.Info("Using campaign from list", "campaignId", campaignID, "name", name)

	return campaignID, nil
}

func createContact(ctx context.Context, conn *getresponse.Connector, campaignID string) (string, error) {
	email := gofakeit.Email()
	name := gofakeit.Name()

	recordData := map[string]any{
		"email": email,
		"name":  name,
		"campaign": map[string]any{
			"campaignId": campaignID,
		},
		"note": fmt.Sprintf("E2E test at %s", time.Now().Format(time.RFC3339)),
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	})
	if err != nil {
		return "", err
	}

	utils.DumpJSON(res, os.Stdout)

	if res.RecordId != "" {
		return res.RecordId, nil
	}

	slog.Warn("No recordId (202 Accepted); finding contact by email")
	return findContactByEmail(ctx, conn, email)
}

func findContactByEmail(ctx context.Context, conn *getresponse.Connector, email string) (string, error) {
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name", "createdOn"),
		Filter:     fmt.Sprintf("query[email]=%s", email),
		PageSize:   10,
	}

	delays := []time.Duration{3 * time.Second, 5 * time.Second, 8 * time.Second}
	for i, delay := range delays {
		time.Sleep(delay)

		res, err := conn.Read(ctx, params)
		if err != nil {
			return "", fmt.Errorf("searching for contact: %w", err)
		}

		if res.Rows == 0 || len(res.Data) == 0 {
			if i < len(delays)-1 {
				slog.Debug("Contact not found yet, retrying", "attempt", i+1)
				continue
			}
			return "", fmt.Errorf("contact not found after creation (tried %d times)", len(delays))
		}

		for _, row := range res.Data {
			if rowEmail, ok := row.Raw["email"].(string); ok && rowEmail == email {
				if contactID, ok := row.Raw["contactId"].(string); ok {
					return contactID, nil
				}
			}
		}

		if i < len(delays)-1 {
			continue
		}
		return "", fmt.Errorf("contactId not in response")
	}

	return "", fmt.Errorf("contact not found")
}

func updateContact(ctx context.Context, conn *getresponse.Connector, contactID string) error {
	recordData := map[string]any{
		"name": gofakeit.Name(),
		"note": fmt.Sprintf("Updated at %s", time.Now().Format(time.RFC3339)),
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: recordData,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if res.RecordId != contactID {
		return fmt.Errorf("expected contactId %s, got %s", contactID, res.RecordId)
	}

	return nil
}

func deleteContact(ctx context.Context, conn *getresponse.Connector, contactID string) error {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)
	if !res.Success {
		return fmt.Errorf("delete reported failure")
	}

	return nil
}
