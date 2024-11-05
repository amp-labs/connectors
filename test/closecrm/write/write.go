package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/closecrm"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := closecrm.GetCloseConnector(ctx)
	defer utils.Close(conn)

	if err := createLead(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateLead(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createContact(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

}

func createLead(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "lead",
		RecordData: map[string]any{
			"name":        "Bluth Company",
			"url":         "http://thebluthcompany.tumblr.com/",
			"description": "Best. Show. Ever.",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateLead(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "lead",
		RecordId:   "lead_UaHMFD5GgwUArEb6eZr21mOhDXkhbEUi9NNxoNkByYC",
		RecordData: map[string]any{
			"url": "http://thebluthcompany.pumblr.com",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createContact(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"name":    "John Smith",
			"title":   "President",
			"lead_id": "lead_UaHMFD5GgwUArEb6eZr21mOhDXkhbEUi9NNxoNkByYC",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
