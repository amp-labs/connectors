package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	googleConn "github.com/amp-labs/connectors/providers/google"
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
		Request: &googleConn.GmailSubscribeRequest{
			TopicName:           "projects/ampersanddev/topics/gmail-event-received",
			LabelIDs:            []string{"INBOX", "UNREAD", "Label_1"}, // system & custom labels.
			LabelFilterBehavior: "include",
		},
	}

	subscribeResult, err := conn.Mail.Subscribe(ctx, subscribeParams)
	if err != nil {
		slog.Error("subscription connector", "subscribe action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	slog.Info("created a subscriber")

	// Exercise history.list using the historyId returned from users.watch().
	// Round-trip through JSON to extract the historyId without reaching into
	// the internal mail package.
	var watchPayload struct {
		HistoryID string `json:"historyId"`
	}

	raw, err := json.Marshal(subscribeResult.Result)
	if err != nil {
		slog.Error("subscription connector", "marshaling watch result", err)
		return
	}

	if err := json.Unmarshal(raw, &watchPayload); err != nil {
		slog.Error("subscription connector", "unmarshaling watch result", err)
		return
	}

	historyResult, err := conn.HistoryList(ctx, googleConn.HistoryListParams{
		StartHistoryID: watchPayload.HistoryID,
	})
	if err != nil {
		slog.Error("subscription connector", "history list action", err)
		return
	}

	slog.Info("history.list succeeded",
		"newCheckpointHistoryId", historyResult.HistoryID,
		"recordCount", len(historyResult.History))

	utils.DumpJSON(historyResult, os.Stdout)

	subscribeResult, err = conn.Mail.RunScheduledMaintenance(ctx, subscribeParams, subscribeResult)
	if err != nil {
		slog.Error("subscription connector", "schedule maintenance action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)

	subscribeResult2, err := conn.Mail.UpdateSubscription(ctx, subscribeParams, subscribeResult)
	if err != nil {
		slog.Error("subscription connector", "update subscription action", err)
		return
	}

	utils.DumpJSON(subscribeResult2, os.Stdout)

	if err := conn.Mail.DeleteSubscription(ctx, *subscribeResult); err != nil {
		slog.Error("subscription connector", "delete action", err)
		return
	}

	utils.DumpJSON(subscribeResult, os.Stdout)
}
