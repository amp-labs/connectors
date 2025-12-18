package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/memstore"
	memstoretest "github.com/amp-labs/connectors/test/memstore"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Set up logging
	utils.SetupLogging()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get configured connector
	conn := memstoretest.GetMemStoreConnector(ctx)

	slog.Info("=== MemStore Write Operations Examples ===")
	fmt.Println()

	// Run test functions
	testCreateContact(ctx, conn)
	testCreateCompany(ctx, conn)
	testCreateDeal(ctx, conn)
	testUpdateContact(ctx, conn)
	testBulkCreate(ctx, conn)

	slog.Info("All write operations completed successfully")
}

// testCreateContact demonstrates creating a contact using GenerateRandomRecord
func testCreateContact(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Creating contact using GenerateRandomRecord")

	// Generate random contact data
	randomContact, err := conn.GenerateRandomRecord("contacts")
	if err != nil {
		slog.Error("Failed to generate random contact", "error", err)
		return
	}

	// Write the contact
	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: randomContact,
	})
	if err != nil {
		slog.Error("Failed to create contact", "error", err)
		return
	}

	slog.Info("Contact created successfully", "recordId", result.RecordId)
	printJSON("Created Contact", randomContact)
	fmt.Println()
}

// testCreateCompany demonstrates creating a company with explicit field values
func testCreateCompany(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Creating company with explicit field values")

	companyData := map[string]any{
		"name":          "Acme Corporation",
		"industry":      "technology",
		"employeeCount": 150,
		"website":       "https://acme.example.com",
	}

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "companies",
		RecordData: companyData,
	})
	if err != nil {
		slog.Error("Failed to create company", "error", err)
		return
	}

	slog.Info("Company created successfully", "recordId", result.RecordId)
	printJSON("Created Company", companyData)
	fmt.Println()
}

// testCreateDeal demonstrates creating a deal with relationships to contacts and companies
func testCreateDeal(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Creating deal with relationships")

	// First, create a contact and company to reference
	contactData := map[string]any{
		"email":     "john.doe@example.com",
		"firstName": "John",
		"lastName":  "Doe",
		"status":    "active",
	}
	contactResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: contactData,
	})
	if err != nil {
		slog.Error("Failed to create contact for deal", "error", err)
		return
	}

	companyData := map[string]any{
		"name":          "Tech Startup Inc",
		"industry":      "technology",
		"employeeCount": 25,
	}
	companyResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "companies",
		RecordData: companyData,
	})
	if err != nil {
		slog.Error("Failed to create company for deal", "error", err)
		return
	}

	// Now create the deal with relationships
	dealData := map[string]any{
		"title":     "Enterprise Software License",
		"amount":    50000.00,
		"stage":     "qualification",
		"contactId": contactResult.RecordId,
		"companyId": companyResult.RecordId,
		"closeDate": "2024-12-31",
	}

	dealResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "deals",
		RecordData: dealData,
	})
	if err != nil {
		slog.Error("Failed to create deal", "error", err)
		return
	}

	slog.Info("Deal created successfully with relationships",
		"dealId", dealResult.RecordId,
		"contactId", contactResult.RecordId,
		"companyId", companyResult.RecordId)
	printJSON("Created Deal", dealData)
	fmt.Println()
}

// testUpdateContact demonstrates updating an existing contact
func testUpdateContact(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Creating and then updating a contact")

	// Create initial contact
	initialData := map[string]any{
		"email":     "jane.smith@example.com",
		"firstName": "Jane",
		"lastName":  "Smith",
		"status":    "active",
		"tags":      []any{"customer", "vip"},
	}

	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: initialData,
	})
	if err != nil {
		slog.Error("Failed to create contact", "error", err)
		return
	}

	slog.Info("Initial contact created", "recordId", createResult.RecordId)
	printJSON("Initial Contact", initialData)

	// Update the contact - change status and add tags
	updateData := map[string]any{
		"status": "inactive",
		"tags":   []any{"customer", "vip", "archived"},
	}

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   createResult.RecordId,
		RecordData: updateData,
	})
	if err != nil {
		slog.Error("Failed to update contact", "error", err)
		return
	}

	slog.Info("Contact updated successfully", "recordId", updateResult.RecordId)
	printJSON("Update Data", updateData)
	fmt.Println()
}

// testBulkCreate demonstrates creating multiple records in a loop
func testBulkCreate(ctx context.Context, conn *memstore.Connector) {
	slog.Info("Creating multiple contacts in bulk")

	contacts := []map[string]any{
		{
			"email":     "alice@example.com",
			"firstName": "Alice",
			"lastName":  "Anderson",
			"status":    "active",
		},
		{
			"email":     "bob@example.com",
			"firstName": "Bob",
			"lastName":  "Brown",
			"status":    "active",
		},
		{
			"email":     "charlie@example.com",
			"firstName": "Charlie",
			"lastName":  "Chen",
			"status":    "inactive",
		},
	}

	recordIds := make([]string, 0, len(contacts))

	for i, contactData := range contacts {
		result, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "contacts",
			RecordData: contactData,
		})
		if err != nil {
			slog.Error("Failed to create contact in bulk", "index", i, "error", err)
			continue
		}
		recordIds = append(recordIds, result.RecordId)
		slog.Info("Bulk contact created", "index", i, "recordId", result.RecordId)
	}

	slog.Info("Bulk creation completed", "totalCreated", len(recordIds))
	printJSON("Created Record IDs", recordIds)
	fmt.Println()
}

// printJSON prints data as formatted JSON
func printJSON(label string, data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal JSON", "error", err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(jsonData))
}
