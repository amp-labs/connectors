package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/goto"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	err := testCreateMeeting(ctx)
	if err != nil {
		fmt.Println("Error creating meeting: ", err)
		os.Exit(1)
	}

	err = testCreateWebhooks(ctx)
	if err != nil {
		fmt.Println("Error creating webhooks: ", err)
		os.Exit(1)
	}

	err = testCreateAttribute(ctx)
	if err != nil {
		fmt.Println("Error creating attribute: ", err)
		os.Exit(1)
	}

	fmt.Println("Tests completed successfully")
}

func testCreateMeeting(ctx context.Context) error {
	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)
	startTime := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	endTime := time.Now().Add(7*24*time.Hour + time.Hour).UTC().Format(time.RFC3339)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "meetings",
		RecordData: map[string]any{
			"subject":            "Test Webinar from Ampersand",
			"starttime":          startTime,
			"endtime":            endTime,
			"passwordrequired":   false,
			"meetingtype":        "immediate",
			"conferencecallinfo": "string",
		},
	})
	if err != nil {
		return fmt.Errorf("error writing to GoTo: %w", err)
	}

	fmt.Println("Created meeting..")
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreateWebhooks(ctx context.Context) error {
	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "webhooks",
		RecordData: []map[string]any{
			{
				"callbackUrl":  "https://webhook.site/4930c4f8-c82c-4183-9fb0-bce3b7450cd1",
				"eventName":    "registrant.joined",
				"eventVersion": "1.0.0",
				"product":      "g2w",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error writing to GoTo: %w", err)
	}

	fmt.Println("Created webhook..")
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreateAttribute(ctx context.Context) error {
	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "attributes",
		RecordData: map[string]any{
			"name": gofakeit.Username(),
		},
	})
	if err != nil {
		return fmt.Errorf("error writing to GoTo: %w", err)
	}

	fmt.Println("Created attribute..")
	utils.DumpJSON(res, os.Stdout)

	return nil
}
