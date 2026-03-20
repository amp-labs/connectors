package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := google.GetGoogleMailConnector(ctx)

	subscribeParams := common.SubscribeParams{
		Request: map[string]any{
			"topicName":           "projects/ampersanddev/topics/gmail-notifications",
			"labelIds":            []string{"INBOX", "UNREAD", "Label_1"}, // system & custom labels.
			"labelFilterBehavior": "include",
		},
	}

	subscribeResult, err := conn.Mail.Subscribe(ctx, subscribeParams)
	if err != nil {
		log.Fatal(err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)
}
