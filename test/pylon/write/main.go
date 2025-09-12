package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/pylon"
	"github.com/amp-labs/connectors/test/pylon"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := pylon.GetConnector(ctx)

	taskId, err := testCreatingTasks(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateTasks(ctx, conn, taskId); err != nil {
		return err
	}

	contactId, err := testCreatingContact(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateContacts(ctx, conn, contactId); err != nil {
		return err
	}

	return nil
}

func testCreatingTasks(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"title": "Ampersand write test",
		},
	}

	slog.Info("Creating tasks...")

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

	return res.Data["id"].(string), nil
}

func testUpdateTasks(ctx context.Context, conn *cc.Connector, taskId string) error {
	params := common.WriteParams{
		ObjectName: "tasks",
		RecordId:   taskId,
		RecordData: map[string]any{
			"title": "Updated ampersand test demo",
		},
	}

	slog.Info("Updating task...")

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

func testCreatingContact(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"name":  "ampersand test",
			"email": "ampersand@example.com",
		},
	}

	slog.Info("Creating contact...")

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

	return res.Data["id"].(string), nil
}

func testUpdateContacts(ctx context.Context, conn *cc.Connector, contactId string) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactId,
		RecordData: map[string]any{
			"name":  "Updated ampersand test demo",
			"email": "updated@example.com",
		},
	}

	slog.Info("Updating contact...")

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
