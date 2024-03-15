package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/proxy"
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/test"
	"golang.org/x/oauth2"
)

const (
	testLineBreak = "\n=============================================\n"
)

func main() { //nolint:funlen
	fmt.Println("Testing Bulkwrite...")

	creds, err := test.GetCreds("../../creds.json")
	if err != nil {
		slog.Error("Error getting creds", "error", err)
		os.Exit(1)
	}

	clientId := creds.ClientId
	clientSecret := creds.ClientSecret
	accessToken := creds.AccessToken
	refreshToken := creds.RefreshToken

	salesforceWorkspace := creds.Workspace

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", salesforceWorkspace),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", salesforceWorkspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	ctx := context.Background()

	proxyConn, err := connectors.NewProxyConnector(
		providers.Salesforce,
		proxy.WithClient(ctx, http.DefaultClient, cfg, tok),
		proxy.WithCatalogSubstitutions(map[string]string{
			salesforce.PlaceholderWorkspace: salesforceWorkspace,
		}),
	)
	if err != nil {
		slog.Error("Error creating proxy connector", "error", err)

		return
	}

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := salesforce.NewConnector(
		salesforce.WithProxyConnector(proxyConn),
	)
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	logs := make([]string, len(testList))

	var wg sync.WaitGroup
	for i, test := range testList {
		wg.Add(1)

		go func(test testRunner, idx int) {
			defer wg.Done()

			log, err := test.fn(ctx, sfc, test.filePath)
			if err != nil {
				logs[idx] = testLineBreak + test.testTitle + testLineBreak + "\n" + err.Error()
			} else {
				logs[idx] = testLineBreak + test.testTitle + testLineBreak + "\n" + log
			}
		}(test, i)
	}

	wg.Wait()

	for _, log := range logs {
		fmt.Println(log)
	}
}

func testBulkWrite(ctx context.Context, sfc *salesforce.Connector, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening '%s': %w", filePath, err)
	}

	res, err := sfc.BulkWrite(ctx, salesforce.BulkWriteParams{
		ObjectName:      "Touchpoint__c",
		ExternalIdField: "external_id__c",
		CSVData:         file,
		Mode:            "upsert",
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

	jobInfo, err := sfc.GetJobInfo(ctx, res.JobId)
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

var testList = []testRunner{
	{
		filePath:  "./touchpoints_20231130.csv",
		testTitle: "Testing Bulkwrite",
		fn:        testBulkWrite,
	},
	{
		filePath:  "./touchpoints_20231130.csv",
		testTitle: "Testing SuccessResults",
		fn:        testGetJobResultsForFile,
	},
	{
		filePath:  "./touchpoints_partial_failure_20231228.csv",
		testTitle: "Testing Partial Failure",
		fn:        testGetJobResultsForFile,
	},
	{
		filePath:  "./touchpoints_complete_failure_20231228.csv",
		testTitle: "Testing Complete Failure",
		fn:        testGetJobResultsForFile,
	},
}

type testRunner struct {
	filePath  string
	testTitle string
	fn        func(ctx context.Context, sfc *salesforce.Connector, filePath string) (string, error)
}

func testGetJobResultsForFile(ctx context.Context, sfc *salesforce.Connector, fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}

	res, err := sfc.BulkWrite(ctx, salesforce.BulkWriteParams{
		ObjectName:      "Touchpoint__c",
		ExternalIdField: "external_id__c",
		CSVData:         file,
		Mode:            "upsert",
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

	jobResults, err := sfc.GetJobResults(ctx, res.JobId)
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
