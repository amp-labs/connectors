package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/justcall"
	testJustCall "github.com/amp-labs/connectors/test/justcall"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testJustCall.GetJustCallConnector(ctx)

	if err := run(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func run(ctx context.Context, conn *justcall.Connector) error {
	contactID, err := testCreateContact(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateContact(ctx, conn, contactID); err != nil {
		return err
	}

	if err := testUpdateContactStatus(ctx, conn, contactID); err != nil {
		return err
	}

	if err := testCreateTag(ctx, conn); err != nil {
		return err
	}

	if err := testUpdateUsersAvailability(ctx, conn); err != nil {
		return err
	}

	if err := testUpdateCall(ctx, conn); err != nil {
		return err
	}

	return nil
}

func testCreateContact(ctx context.Context, conn *justcall.Connector) (string, error) {
	phoneNumber := fmt.Sprintf("+1415555%04d", time.Now().Unix()%10000)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"first_name":     "Test",
			"last_name":      "WriteAPI",
			"contact_number": phoneNumber,
			"email":          "testwrite@example.com",
			"company":        "Test Company",
		},
	})
	if err != nil {
		return "", fmt.Errorf("contacts create: %w", err)
	}

	printResult("contacts (CREATE)", res)

	return res.RecordId, nil
}

func testUpdateContact(ctx context.Context, conn *justcall.Connector, contactID string) error {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: map[string]any{
			"id":         contactID,
			"first_name": "Updated",
			"last_name":  "Contact",
			"company":    "Updated Company",
		},
	})
	if err != nil {
		return fmt.Errorf("contacts update: %w", err)
	}

	printResult("contacts (UPDATE)", res)

	return nil
}

func testUpdateContactStatus(ctx context.Context, conn *justcall.Connector, contactID string) error {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts/status",
		RecordData: map[string]any{
			"id":     contactID,
			"add_to": []string{"dnd"},
		},
	})
	if err != nil {
		return fmt.Errorf("contacts/status update: %w", err)
	}

	printResult("contacts/status (UPDATE)", res)

	return nil
}

func testCreateTag(ctx context.Context, conn *justcall.Connector) error {
	tagName := fmt.Sprintf("TestTag%d", time.Now().Unix()%10000)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"name":       tagName,
			"color_code": "#FF5733",
		},
	})
	if err != nil {
		return fmt.Errorf("tags create: %w", err)
	}

	printResult("tags (CREATE)", res)

	return nil
}

func testUpdateUsersAvailability(ctx context.Context, conn *justcall.Connector) error {
	readRes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name"),
	})
	if err != nil {
		return fmt.Errorf("read users: %w", err)
	}

	if len(readRes.Data) == 0 {
		return fmt.Errorf("no users found")
	}

	agentID := readRes.Data[0].Raw["id"]

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "users/availability",
		RecordData: map[string]any{
			"agent_id":     agentID,
			"is_available": true,
		},
	})
	if err != nil {
		return fmt.Errorf("users/availability update: %w", err)
	}

	printResult("users/availability (UPDATE)", res)

	return nil
}

func testUpdateCall(ctx context.Context, conn *justcall.Connector) error {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "calls",
		RecordId:   "328951212",
		RecordData: map[string]any{
			"notes":  "Updated via API test",
			"rating": "5",
		},
	})
	if err != nil {
		return fmt.Errorf("calls update: %w", err)
	}

	printResult("calls (UPDATE)", res)

	return nil
}

func printResult(name string, res *common.WriteResult) {
	jsonStr, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("\n=== %s ===\n", name)
	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")
}
