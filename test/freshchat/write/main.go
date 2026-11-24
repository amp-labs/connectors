package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors/common"
	fc "github.com/amp-labs/connectors/providers/freshchat"
	"github.com/amp-labs/connectors/test/freshchat"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := freshchat.NewConnector(ctx)

	userID, err := testCreatingUsers(ctx, conn)
	if err != nil {
		return err
	}

	return testUpdatingUsers(ctx, conn, userID)
}

// testCreatingUsers creates a user and returns its identifier.
func testCreatingUsers(ctx context.Context, conn *fc.Connector) (string, error) {
	// Freshchat rejects a second user with the same email or reference, so each run is unique.
	stamp := time.Now().Unix()

	params := common.WriteParams{
		ObjectName: "users",
		RecordData: map[string]any{
			"avatar": map[string]any{
				"url": "https://web.freshchat.com/img/johndoe.png",
			},
			"email":        fmt.Sprintf("milton.doe+%d@mail.com", stamp),
			"first_name":   "Milton",
			"last_name":    "Doe",
			"phone":        "235689714",
			"reference_id": fmt.Sprintf("milton@doe-%d", stamp),
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.RecordId, nil
}

func testUpdatingUsers(ctx context.Context, conn *fc.Connector, userID string) error {
	params := common.WriteParams{
		ObjectName: "users",
		RecordId:   userID,
		RecordData: map[string]any{
			"avatar": map[string]any{
				"url": "https://web.freshchat.com/img/johndoe2.png",
			},
			"timezone": "Africa/nairobi",
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
