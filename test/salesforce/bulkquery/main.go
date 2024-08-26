package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
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

	query := "SELECT Id, Name FROM Account LIMIT 10"

	res, err := sfc.BulkQuery(ctx, query)
	if err != nil {
		slog.Error("Error querying", "error", err)
		return
	}

	slog.Info("Query Info")
	prettyPrint(res)

	_, err = getResultInLoop(ctx, sfc, res.Id)
	if err != nil {
		slog.Error("Error getting job results", "error", err)
		return
	}

	slog.Info("Job completed... fetching results")

	// Get the results
	result, err := sfc.GetBulkQueryResults(ctx, res.Id)
	if err != nil {
		slog.Error("Error getting query results", "error", err)
		return
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		slog.Error("Error reading response body", "error", err)
		return
	}

	defer func() {
		if res != nil && result.Body != nil {
			if closeErr := result.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	slog.Info("Query results")
	fmt.Println(string(body))
}

func getResultInLoop(ctx context.Context, sfc *salesforce2.Connector, jobId string) (*salesforce2.GetJobInfoResult, error) {
	done := false
	var jobRes *salesforce2.GetJobInfoResult
	var err error

	for !done {
		fmt.Print(".")
		time.Sleep(2 * time.Second)

		jobRes, err = sfc.GetBulkQueryInfo(ctx, jobId)
		if err != nil {
			return nil, fmt.Errorf("error getting job results: %w", err)
		}

		done = isJobDone(jobRes)
	}
	fmt.Println(".")

	return jobRes, nil
}

func isJobDone(jobRes *salesforce2.GetJobInfoResult) bool {
	return jobRes.State == salesforce2.JobStateComplete || jobRes.State == salesforce2.JobStateFailed || jobRes.State == salesforce2.JobStateAborted
}

func prettyPrint(s any) {
	fmt.Println(utils.PrettyFormatStruct(s))
}
