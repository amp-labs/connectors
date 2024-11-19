package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/salesforce/bulk"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/tools/fileconv"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	// We first create objects in Salesforce,
	// and then we generate an in-memory CSV of the Salesforce IDs of the newly created objects,
	// so that we can bulk-delete them.
	// For convenience the same records are used that were created by write operation.
	// This is sort of a cleanup with a demonstration of BulkDelete.
	objectCSVToDelete, err := createObjects(ctx, conn,
		fileconv.NewSiblingFileLocator().AbsPathTo("../write/opportunities.csv"),
	)
	if err != nil {
		slog.Error("Error creating file to delete", "error", err)
		return
	}

	deleteRes, err := conn.BulkDelete(ctx, salesforce.BulkOperationParams{
		ObjectName: "Opportunity",
		CSVData:    bytes.NewReader(objectCSVToDelete),
	})
	if err != nil {
		slog.Error("Error bulk deleting", "error", err)
		return
	}

	slog.Info("Bulk delete job created", "res", deleteRes)

	// Get delete results. waits for the job to complete
	deleteResult, err := bulk.GetResultInLoop(ctx, conn, deleteRes.JobId)
	if err != nil {
		slog.Error("Error getting bulk delete job results", "error", err)
		return
	}

	slog.Info("Bulk delete job done")

	utils.DumpJSON(deleteResult, os.Stdout)
}

func createObjects(ctx context.Context, conn *salesforce.Connector, filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening '%s': %w", filePath, err)
	}
	defer utils.Close(file)

	// Write the records to Salesforce, so that we can delete them later.
	writeRes, err := conn.BulkWrite(ctx, salesforce.BulkOperationParams{
		ObjectName:      "Opportunity",
		ExternalIdField: "external_id__c",
		CSVData:         file,
		Mode:            salesforce.UpsertMode,
	})
	if err != nil {
		return nil, fmt.Errorf("error bulk writing to prepare bulk delete: %w", err)
	}

	slog.Info("Preparing objects to delete", "res", writeRes)

	// wait for the job to complete
	_, err = bulk.GetResultInLoop(ctx, conn, writeRes.JobId)
	if err != nil {
		return nil, fmt.Errorf("error getting bulk write job results: %w", err)
	}

	slog.Info("Records created, now deleting them.")

	return bulk.GetRecordIDsForJob(ctx, conn, writeRes.JobId)
}
