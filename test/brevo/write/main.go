package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/brevo"
	hs "github.com/amp-labs/connectors/test/brevo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetBrevoConnector(ctx)

	if err := createSmtpEmail(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createSmtpTemplates(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := AddBlockedDomains(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func createSmtpEmail(ctx context.Context, conn *brevo.Connector) error {
	config := common.WriteParams{
		ObjectName: "smtp/email",
		RecordData: map[string]any{
			"subject":     "Login Email confirmation",
			"htmlContent": "<!DOCTYPE html> <html> <body> <h1>This is sample HTML</h1> </html>",
			"sender": map[string]string{
				"name":  "Mary from MyShop",
				"email": "no-reply@myshop.com",
			},
			"to": []map[string]any{
				{
					"email": "dipu@withampersand.com",
					"name":  "this is a test lol",
				},
			},
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

func createSmtpTemplates(ctx context.Context, conn *brevo.Connector) error {
	config := common.WriteParams{
		ObjectName: "smtp/templates",
		RecordData: map[string]any{
			"sender": map[string]string{
				"name":  "This is test",
				"email": "integration.user+Brevo@withampersand.com",
			},
			"subject":      "Thanks for your purchase !",
			"templateName": "Order Confirmation - EN",
			"htmlContent":  "The order n°xxxxx has been confirmed. Thanks for your purchase",
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

func AddBlockedDomains(ctx context.Context, conn *brevo.Connector) error {
	config := common.WriteParams{
		ObjectName: "smtp/blockedDomains",
		RecordData: map[string]any{
			"domain": "chaurfdfdasdfdiy.com",
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
