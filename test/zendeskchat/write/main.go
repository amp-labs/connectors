package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	zc "github.com/amp-labs/connectors/providers/zendeskchat"
	"github.com/amp-labs/connectors/test/zendeskchat"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := zendeskchat.GetConnector(ctx)

	err := testCreatingShortcuts(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingTriggers(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingChat(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingShortcuts(ctx context.Context, conn *zc.Connector) error {
	params := common.WriteParams{
		ObjectName: "shortcuts",
		RecordData: map[string]any{
			"name":    "intro",
			"message": "Hi! Do you need assistance?",
			"options": "Yes/No",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreatingTriggers(ctx context.Context, conn *zc.Connector) error {
	params := common.WriteParams{
		ObjectName: "triggers",
		RecordData: map[string]any{
			"name":        "Test Trigger Z",
			"enabled":     1,
			"description": "Visitor requested chat",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreatingChat(ctx context.Context, conn *zc.Connector) error {
	params := common.WriteParams{
		ObjectName: "chats",
		RecordData: map[string]any{
			"visitor": map[string]any{
				"phone": "",
				"notes": "",
				"id":    "1.12345",
				"name":  "John",
				"email": "visitor_john@doe.com",
			},
			"message":   "Hi there!",
			"type":      "offline_msg",
			"timestamp": 1444156010,
			"session": map[string]any{
				"browser":      "Safari",
				"city":         "Orlando",
				"country_code": "US",
				"country_name": "United States",
				"end_date":     "2014-10-09T05:46:47Z",
				"id":           "141109.654464.1KhqS0Nw",
				"ip":           "67.32.299.96",
				"platform":     "Mac OS",
				"region":       "Florida",
				"start_date":   "2014-10-09T05:28:31Z",
				"user_agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10) AppleWebKit/600.1.25 (KHTML, like Gecko) Version/8.0 Safari/600.1.25",
			},
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
