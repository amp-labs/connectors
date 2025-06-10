package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/helpscout"
	hs "github.com/amp-labs/connectors/test/helpscout"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetHelpScoutConnector(ctx)

	if err := createConversations(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createCustomers(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateConversations(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func createConversations(ctx context.Context, conn *helpscout.Connector) error {
	config := common.WriteParams{
		ObjectName: "conversations",
		RecordData: map[string]any{
			"subject": "Subject",
			"customer": map[string]string{
				"email":     "bear@acme.com",
				"firstName": "Vernon",
				"lastName":  "Bear",
			},
			"mailboxId": 339541,
			"type":      "email",
			"status":    "active",
			"createdAt": "2012-10-10T12:00:00Z",
			"threads": []map[string]any{
				{"type": "customer",
					"customer": map[string]any{
						"email": "bear@acme.com",
					},
					"text": "Hello, Help Scout. How are you?"},
			}},
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

func createCustomers(ctx context.Context, conn *helpscout.Connector) error {
	config := common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"firstName":    "Vernon",
			"lastName":     "Bear",
			"photoUrl":     "https://api.helpscout.net/img/some-avatar.jpg",
			"photoType":    "twitter",
			"jobTitle":     "CEO and Co-Founder",
			"location":     "Greater Dallas/FT Worth Area",
			"background":   "I've worked with Vernon before and he's really great.",
			"age":          "30-35",
			"gender":       "Male",
			"organization": "Acme, Inc",
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

func updateConversations(ctx context.Context, conn *helpscout.Connector) error {
	config := common.WriteParams{
		ObjectName: "conversations",
		RecordId:   "2855294683",
		RecordData: map[string]any{
			"op":    "replace",
			"path":  "/subject",
			"value": "super cool new subject",
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
