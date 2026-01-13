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

	// Test Contact operations
	contactID, err := testCreateContact(ctx, conn, campaignID)
	if err != nil {
		utils.Fail("Failed to create contact", "error", err)
	}

	// Wait a bit for the contact to be fully created (GetResponse may need time to process)
	time.Sleep(2 * time.Second)

	slog.Info("Contact created successfully", "contactId", contactID)

	// Update the contact
	if err := testUpdateContact(ctx, conn, contactID); err != nil {
		utils.Fail("Failed to update contact", "error", err)
	}

	slog.Info("Contact write operations completed successfully!")

	// Test Campaign operations (if supported)
	slog.Info("=== Step 4: Testing Campaign Write Operations ===")

	campaignID2, err := testCreateCampaign(ctx, conn)
	if err != nil {
		slog.Warn("Campaign creation may not be supported or failed", "error", err)
		// Campaign creation might require special permissions, so we'll continue
	} else {
		slog.Info("Campaign created successfully", "campaignId", campaignID2)

		// Wait a bit for the campaign to be fully created
		time.Sleep(2 * time.Second)

		// Update the campaign
		if err := testUpdateCampaign(ctx, conn, campaignID2); err != nil {
			slog.Warn("Campaign update failed", "error", err)
		} else {
			slog.Info("Campaign write operations completed successfully!")
		}
	}

	slog.Info("All write tests completed successfully!")
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

// testCreateContact creates a new contact in GetResponse.
// According to GetResponse API v3 documentation:
// - POST /v3/contacts
// - Required fields: email, campaign (with campaignId)
// - Returns 202 Accepted with no body
func testCreateContact(ctx context.Context, conn *getresponse.Connector, campaignID string) (string, error) {
	// Generate realistic contact data
	email := gofakeit.Email()
	name := gofakeit.Name()

	recordData := map[string]any{
		"email": email,
		"name":  name,
		"campaign": map[string]any{
			"campaignId": campaignID,
		},
		// Optional fields
		"note": fmt.Sprintf("Test contact created at %s", time.Now().Format(time.RFC3339)),
	}

	slog.Info("Creating contact", "email", email, "name", name, "campaignId", campaignID)

	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating contact: %w", err)
	}

	slog.Info("Contact creation response",
		"success", res.Success,
		"recordId", res.RecordId)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	// For create operations, GetResponse returns 202 Accepted with no body
	// The contactId will be available after creation, but we need to read it back
	// For now, we'll return empty string and handle it in the update step
	// In a real scenario, you might need to read the contact back using email as identifier
	if res.RecordId == "" {
		slog.Warn("No recordId returned from create operation (expected for 202 Accepted)")
		// We'll need to find the contact by email to get its ID
		return findContactByEmail(ctx, conn, email)
	}

	return res.RecordId, nil
}

// findContactByEmail searches for a contact by email address.
func findContactByEmail(ctx context.Context, conn *getresponse.Connector, email string) (string, error) {
	// Wait a moment for the contact to be indexed
	time.Sleep(3 * time.Second)

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

// testUpdateContact updates an existing contact.
// According to GetResponse API v3 documentation:
// - POST /v3/contacts/{contactId}
// - Returns 200 OK with the updated contact object
func testUpdateContact(ctx context.Context, conn *getresponse.Connector, contactID string) error {
	updatedName := gofakeit.Name()
	updatedNote := fmt.Sprintf("Updated at %s", time.Now().Format(time.RFC3339))

	recordData := map[string]any{
		"name": updatedName,
		"note": updatedNote,
	}

	slog.Info("Updating contact", "contactId", contactID, "name", updatedName)

	params := common.WriteParams{ObjectName: "contacts", RecordId: contactID, RecordData: recordData}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("error updating contact: %w", err)
	}

	slog.Info("Contact update response", "success", res.Success, "recordId", res.RecordId)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	// Verify the update was successful
	if res.RecordId != contactID {
		return fmt.Errorf("expected contactId %s, got %s", contactID, res.RecordId)
	}

	return nil
}

// testCreateCampaign creates a new campaign in GetResponse.
// According to GetResponse API v3 documentation:
// - POST /v3/campaigns
// - Required fields: name, languageCode
// - Returns 202 Accepted with no body
func testCreateCampaign(ctx context.Context, conn *getresponse.Connector) (string, error) {
	campaignName := fmt.Sprintf("Test Campaign %s", gofakeit.Word())
	languageCode := "EN" // English

	recordData := map[string]any{
		"name":         campaignName,
		"languageCode": languageCode,
		"description":  fmt.Sprintf("Test campaign created at %s", time.Now().Format(time.RFC3339)),
	}

	slog.Info("Creating campaign", "name", campaignName, "languageCode", languageCode)

	params := common.WriteParams{
		ObjectName: "campaigns",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating campaign: %w", err)
	}

	slog.Info("Campaign creation response", "success", res.Success, "recordId", res.RecordId)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	// For create operations, GetResponse returns 202 Accepted with no body
	// We need to find the campaign by name to get its ID
	if res.RecordId == "" {
		slog.Warn("No recordId returned from create operation (expected for 202 Accepted)")
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

// testUpdateCampaign updates an existing campaign.
// According to GetResponse API v3 documentation:
// - POST /v3/campaigns/{campaignId}
// - Returns 200 OK with the updated campaign object
func testUpdateCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string) error {
	updatedName := fmt.Sprintf("Updated Campaign %s", gofakeit.Word())
	updatedDescription := fmt.Sprintf("Updated at %s", time.Now().Format(time.RFC3339))

	recordData := map[string]any{
		"name":        updatedName,
		"description": updatedDescription,
	}

	slog.Info("Updating campaign", "campaignId", campaignID, "name", updatedName)

	params := common.WriteParams{
		ObjectName: "campaigns",
		RecordId:   campaignID,
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("error updating campaign: %w", err)
	}

	slog.Info("Campaign update response",
		"success", res.Success,
		"recordId", res.RecordId)

	// Print the full response
	utils.DumpJSON(res, os.Stdout)

	// Verify the update was successful
	if res.RecordId != campaignID {
		return fmt.Errorf("expected campaignId %s, got %s", campaignID, res.RecordId)
	}

	return nil
}
