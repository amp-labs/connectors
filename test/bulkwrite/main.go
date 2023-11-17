package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
)

func main() { //nolint:funlen
	file, err := os.Open("../../creds.json")
	if err != nil {
		slog.Error("Error opening creds.json", "error", err)

		return
	}

	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Error reading creds.json", "error", err)

		return
	}

	var credsMap map[string]string

	if err := json.Unmarshal(byteValue, &credsMap); err != nil {
		slog.Error("Error marshalling creds.json", "error", err)

		return
	}

	clientId := credsMap["CLIENT_ID"]
	clientSecret := credsMap["CLIENT_SECRET"]
	accessToken := credsMap["ACCESS_TOKEN"]
	refreshToken := credsMap["REFRESH_TOKEN"]
	salesforceSubdomain := credsMap["SALESFORCE_SUBDOMAIN"]

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", salesforceSubdomain),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", salesforceSubdomain),
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

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithSubdomain(salesforceSubdomain))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	res, err := sfc.BulkWrite(ctx, salesforce.BulkWriteParams{
		ObjectName: "Touchpoint__c",
		ExternalId: "external_id__c",
		FilePath:   "../../playground/bulkapi/touchpoints.csv",
	})
	if err != nil {
		slog.Error("Error bulk writing", "error", err)

		return
	}

	bulkRes, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		slog.Error("Error marshalling bulk result", "error", err)
	}

	fmt.Println("Upload complete.")
	fmt.Println(string(bulkRes))

	time.Sleep(5 * time.Second)

	jobInfo, err := sfc.GetJobInfo(ctx, res.JobId)
	if err != nil {
		slog.Error("Error getting job result", "error", err)

		return
	}

	jsonData, err := json.MarshalIndent(jobInfo, "", "    ")
	if err != nil {
		slog.Error("Error marshalling job result", "error", err)
	}

	fmt.Println("Write Result")
	fmt.Println(string(jsonData))
}
