package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/connectors/providers/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func testBulkWriteOpportunity(ctx context.Context, conn *salesforce.Connector, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening '%s': %w", filePath, err)
	}
	defer testUtils.Close(file)

	res, err := conn.BulkWrite(ctx, salesforce.BulkOperationParams{
		ObjectName:      "Opportunity",
		ExternalIdField: "external_id__c",
		CSVData:         file,
		Mode:            salesforce.UpsertMode,
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

	time.Sleep(5 * time.Second)

	jobInfo, err := conn.GetJobInfo(ctx, res.JobId)
	if err != nil {
		return "", fmt.Errorf("error getting job info: %w", err)
	}

	jsonData, err := json.MarshalIndent(jobInfo, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error marshalling job info: %w", err)
	}

	log += "Write Result\n"
	log += string(jsonData) + "\n"

	return log, nil
}
