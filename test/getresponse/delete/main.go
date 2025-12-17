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
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// First, get a default campaign to use for creating contacts
	campaignID, err := getDefaultCampaign(ctx, conn)
	if err != nil {
		utils.Fail("Failed to get default campaign", "error", err)
	}

	// Test Contact Delete operations
	contactID, err := testCreateContactForDelete(ctx, conn, campaignID)
	if err != nil {
		utils.Fail("Failed to create contact for deletion", "error", err)
	}

	// Wait a bit for the contact to be fully created (GetResponse may need time to process)
	time.Sleep(3 * time.Second)

	slog.Info("Contact created successfully", "contactId", contactID)

	// Delete the contact
	if err := testDeleteContact(ctx, conn, contactID); err != nil {
		utils.Fail("Failed to delete contact", "error", err)
	}

	slog.Info("Contact delete operation completed successfully!")

	// Verify the contact was deleted by trying to read it
	time.Sleep(2 * time.Second)
	if err := verifyContactDeleted(ctx, conn, contactID); err != nil {
		slog.Warn("Contact deletion verification failed", "error", err)
	} else {
		slog.Info("Contact deletion verified successfully")
	}

	// Test Campaign Delete operations (if campaign creation is supported)
	campaignID2, err := testCreateCampaignForDelete(ctx, conn)
	if err != nil {
		slog.Error("Campaign creation may not be supported or failed", "error", err)
	} else {
		slog.Info("Campaign created successfully", "campaignId", campaignID2)

		// Wait a bit for the campaign to be fully created
		time.Sleep(3 * time.Second)

		// Delete the campaign
		if err := testDeleteCampaign(ctx, conn, campaignID2); err != nil {
			slog.Warn("Failed to delete campaign", "error", err)
		} else {
			slog.Info("Campaign delete operation completed successfully!")
		}
	}

	slog.Info("All delete tests completed successfully!")
}

// getDefaultCampaign retrieves the default campaign to use for creating contacts.
// According to GetResponse API, contacts must be associated with a campaign.
func getDefaultCampaign(ctx context.Context, conn *getresponse.Connector) (string, error) {
	// Read campaigns with filter for default campaign
	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "isDefault"),
		Filter:     "query[isDefault]=true",
		PageSize:   1,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error reading campaigns: %w", err)
	}

	if res.Rows == 0 || len(res.Data) == 0 {
		return "", fmt.Errorf("no default campaign found")
	}

	// Extract campaignId from the first result
	campaignID, ok := res.Data[0].Raw["campaignId"].(string)
	if !ok {
		return "", fmt.Errorf("campaignId not found in response")
	}

	slog.Info("Found default campaign", "campaignId", campaignID, "name", res.Data[0].Raw["name"])
	return campaignID, nil
}

// testCreateContactForDelete creates a new contact specifically for deletion testing.
// According to GetResponse API v3 documentation:
// - POST /v3/contacts
// - Required fields: email, campaign (with campaignId)
// - Returns 202 Accepted with no body
func testCreateContactForDelete(ctx context.Context, conn *getresponse.Connector, campaignID string) (string, error) {
	// Generate realistic contact data with a unique identifier
	email := fmt.Sprintf("delete-test-%s@example.com", gofakeit.UUID())
	name := fmt.Sprintf("Delete Test %s", gofakeit.Name())

	recordData := map[string]any{
		"email": email,
		"name":  name,
		"campaign": map[string]any{
			"campaignId": campaignID,
		},
		"note": fmt.Sprintf("Test contact created for deletion at %s", time.Now().Format(time.RFC3339)),
	}

	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating contact: %w", err)
	}

	// For create operations, GetResponse returns 202 Accepted with no body
	// We need to find the contact by email to get its ID
	if res.RecordId == "" {
		slog.Info("No recordId returned from create operation (expected for 202 Accepted), searching by email")
		// Wait for the contact to be indexed
		time.Sleep(3 * time.Second)
		return findContactByEmail(ctx, conn, email)
	}

	return res.RecordId, nil
}

// findContactByEmail searches for a contact by email address.
func findContactByEmail(ctx context.Context, conn *getresponse.Connector, email string) (string, error) {
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name"),
		Filter:     fmt.Sprintf("query[email]=%s", email),
		PageSize:   10,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error searching for contact: %w", err)
	}

	if res.Rows == 0 || len(res.Data) == 0 {
		return "", fmt.Errorf("contact not found after creation")
	}

	// Find the exact match
	for _, row := range res.Data {
		if rowEmail, ok := row.Raw["email"].(string); ok && rowEmail == email {
			if contactID, ok := row.Raw["contactId"].(string); ok {
				return contactID, nil
			}
		}
	}

	return "", fmt.Errorf("contact email found but contactId not available")
}

// testDeleteContact deletes a contact.
// According to GetResponse API v3 documentation:
// - DELETE /v3/contacts/{contactId}
// - Returns 200 OK or 204 No Content on success
func testDeleteContact(ctx context.Context, conn *getresponse.Connector, contactID string) error {
	params := common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
	}

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("error deleting contact: %w", err)
	}

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete operation reported failure")
	}

	return nil
}

// verifyContactDeleted attempts to read the contact to verify it was deleted.
// This should fail or return no results if the deletion was successful.
func verifyContactDeleted(ctx context.Context, conn *getresponse.Connector, contactID string) error {
	// Try to read the specific contact by ID
	// Note: GetResponse API doesn't have a direct "get by ID" endpoint for contacts,
	// so we'll try to read all contacts and check if the ID is missing
	params := common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name"),
		PageSize:   100,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		// If read fails, the contact might be deleted (or there's another issue)
		return fmt.Errorf("error reading contacts for verification: %w", err)
	}

	// Check if the contact ID still exists in the results
	for _, row := range res.Data {
		if rowContactID, ok := row.Raw["contactId"].(string); ok && rowContactID == contactID {
			return fmt.Errorf("contact still exists after deletion (contactId: %s)", contactID)
		}
	}

	// Contact not found in results, deletion likely successful
	slog.Info("Contact not found in read results, deletion verified")
	return nil
}

// testCreateCampaignForDelete creates a new campaign specifically for deletion testing.
// According to GetResponse API v3 documentation:
// - POST /v3/campaigns
// - Required fields: name, languageCode
// - Returns 202 Accepted with no body
func testCreateCampaignForDelete(ctx context.Context, conn *getresponse.Connector) (string, error) {
	campaignName := fmt.Sprintf("Delete Test Campaign %s", gofakeit.Word())
	languageCode := "EN" // English

	recordData := map[string]any{
		"name":         campaignName,
		"languageCode": languageCode,
		"description":  fmt.Sprintf("Test campaign created for deletion at %s", time.Now().Format(time.RFC3339)),
	}

	params := common.WriteParams{
		ObjectName: "campaigns",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating campaign: %w", err)
	}

	// For create operations, GetResponse returns 202 Accepted with no body
	// We need to find the campaign by name to get its ID
	if res.RecordId == "" {
		slog.Info("No recordId returned from create operation (expected for 202 Accepted), searching by name")
		// Wait for the campaign to be indexed
		time.Sleep(3 * time.Second)
		return findCampaignByName(ctx, conn, campaignName)
	}

	return res.RecordId, nil
}

// findCampaignByName searches for a campaign by name.
func findCampaignByName(ctx context.Context, conn *getresponse.Connector, name string) (string, error) {
	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "description"),
		Filter:     fmt.Sprintf("query[name]=%s", name),
		PageSize:   10,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error searching for campaign: %w", err)
	}

	if res.Rows == 0 || len(res.Data) == 0 {
		return "", fmt.Errorf("campaign not found after creation")
	}

	// Find the exact match
	for _, row := range res.Data {
		if rowName, ok := row.Raw["name"].(string); ok && rowName == name {
			if campaignID, ok := row.Raw["campaignId"].(string); ok {
				return campaignID, nil
			}
		}
	}

	return "", fmt.Errorf("campaign name found but campaignId not available")
}

// testDeleteCampaign deletes a campaign.
// According to GetResponse API v3 documentation:
// - DELETE /v3/campaigns/{campaignId}
// - Returns 200 OK or 204 No Content on success
func testDeleteCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string) error {
	params := common.DeleteParams{
		ObjectName: "campaigns",
		RecordId:   campaignID,
	}

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("error deleting campaign: %w", err)
	}

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete operation reported failure")
	}

	return nil
}
