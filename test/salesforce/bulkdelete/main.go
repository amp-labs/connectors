package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	salesforce2 "github.com/amp-labs/connectors/providers/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx := context.Background()

	sfc, err := testUtils.Connector(ctx)
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	// We first create objects in Salesforce,
	// and then we generate an in-memory CSV of the Salesforce IDs of the newly created objects,
	// so that we can bulk-delete them."
	objectCSVToDelete, err := prepareObjectsToDelete(ctx, sfc)
	if err != nil {
		slog.Error("Error creating file to delete", "error", err)
		return
	}

	deleteRes, err := sfc.BulkDelete(ctx, salesforce2.BulkOperationParams{
		ObjectName: "Touchpoint__c",
		CSVData:    bytes.NewReader(objectCSVToDelete),
	})
	if err != nil {
		slog.Error("Error bulk deleting", "error", err)
		return
	}

	slog.Info("Bulk delete job created", "res", deleteRes)

	// Get delete results. waits for the job to complete
	deleteResult, err := getResultInLoop(ctx, sfc, deleteRes.JobId)
	if err != nil {
		slog.Error("Error getting bulk delete job results", "error", err)
		return
	}

	slog.Info("Bulk delete job done")

	prettyPrint(deleteResult)
}

func isJobDone(jobRes *salesforce2.JobResults) bool {
	return jobRes.State == salesforce2.JobStateComplete || jobRes.State == salesforce2.JobStateFailed || jobRes.State == salesforce2.JobStateAborted
}

func prettyPrint(s any) {
	fmt.Println(utils.PrettyFormatStruct(s))
}

func getResultInLoop(ctx context.Context, sfc *salesforce2.Connector, jobId string) (*salesforce2.JobResults, error) {
	done := false
	var jobRes *salesforce2.JobResults
	var err error

	for !done {
		fmt.Print(".")
		time.Sleep(2 * time.Second)

		jobRes, err = sfc.GetJobResults(ctx, jobId)
		if err != nil {
			return nil, fmt.Errorf("error getting job results: %w", err)
		}

		done = isJobDone(jobRes)
	}

	return jobRes, nil
}

func csvBytesToSlice(b []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewBuffer(b))

	records := make([][]string, 0)

	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV headers: %w", err)
	}

	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		records = append(records, row)
	}

	return headers, records, nil
}

func prepareObjectsToDelete(ctx context.Context, sfc *salesforce2.Connector) ([]byte, error) {
	testFilePath := "./test/salesforce/bulkdelete/touchpoints_for_bulkdelete_20240325.csv"

	fileToWrite, err := os.Open(testFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer func() {
		if fileToWrite != nil {
			if closeErr := fileToWrite.Close(); closeErr != nil {
				slog.Warn("unable to close file", "error", closeErr)
			}
		}
	}()

	// Write the records to Salesforce, so that we can delete them later.
	writeRes, err := sfc.BulkWrite(ctx, salesforce2.BulkOperationParams{
		ObjectName:      "Touchpoint__c",
		ExternalIdField: "external_id__c",
		CSVData:         fileToWrite,
		Mode:            "upsert",
	})
	if err != nil {
		return nil, fmt.Errorf("error bulk writing to prepare bulk delete: %w", err)
	}

	slog.Info("Preparing objects to delete", "res", writeRes)

	// wait for the job to complete
	_, err = getResultInLoop(ctx, sfc, writeRes.JobId)
	if err != nil {
		return nil, fmt.Errorf("error getting bulk write job results: %w", err)
	}

	slog.Info("Records created, now deleting them.")

	return getIdsFromJobToDelete(ctx, sfc, writeRes.JobId)
}

func getIdsFromJobToDelete(ctx context.Context, sfc *salesforce2.Connector, jobId string) ([]byte, error) {
	// Get the successful results to get the ids to use for the delete
	successRes, err := sfc.GetSuccessfulJobResults(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("error getting successfult write results: %w", err)
	}

	successBody, err := io.ReadAll(successRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading success results body: %w", err)
	}

	defer func() {
		if successRes != nil && successRes.Body != nil {
			if closeErr := successRes.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	headers, rows, err := csvBytesToSlice(successBody)
	if err != nil {
		return nil, fmt.Errorf("error parsing CSV: %w", err)
	}

	// remove all columns except the id column
	csvRecords, err := filterIds(headers, rows)
	if err != nil {
		return nil, fmt.Errorf("error filtering ids: %w", err)
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	w := csv.NewWriter(buf)

	for _, record := range csvRecords {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatalln(err)
	}

	return buf.Bytes(), nil
}

func filterIds(headers []string, rows [][]string) ([][]string, error) {
	// filter ids
	idIndex := -1

	for i, header := range headers {
		if header == "sf__Id" {
			idIndex = i
			break
		}
	}

	if idIndex == -1 {
		return nil, fmt.Errorf("sf__Id not found in successfulResults headers")
	}

	csvRecords := [][]string{{"id"}}

	for _, row := range rows {
		csvRecords = append(csvRecords, []string{strings.Trim(row[idIndex], " ")})
	}

	return csvRecords, nil
}
