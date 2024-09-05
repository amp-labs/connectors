package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/amp-labs/connectors/providers/salesforce"
	"log/slog"
	"os"
	"time"
)

func testGetJobResultsForFile(ctx context.Context, conn *salesforce.Connector, fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}

	res, err := conn.BulkWrite(ctx, salesforce.BulkOperationParams{
		ObjectName:      "Opportunity",
		ExternalIdField: "external_id__c",
		CSVData:         file,
		Mode:            salesforce.Upsert,
	})
	if err != nil {
		return "", fmt.Errorf("error bulk writing: %w", err)
	}

	bulkRes, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error marshalling bulk result: %w", err)
	}

	log := ""

	log += "Upload complete.\n"
	log += string(bulkRes) + "\n"

	time.Sleep(10 * time.Second)

	jobResults, err := conn.GetJobResults(ctx, res.JobId)
	if err != nil {
		return "", fmt.Errorf("error getting job result: %w", err)
	}

	jsonData, err := json.MarshalIndent(jobResults, "", "    ")
	if err != nil {
		slog.Error("Error marshalling job result", "error", err)
		return "", fmt.Errorf("error marshalling job result: %w", err)
	}

	log += "Write Result\n"
	log += string(jsonData) + "\n"

	return log, nil
}
