package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/asana"
	"github.com/amp-labs/connectors/test/asana"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := asana.GetAsanaConnector(ctx)

	if err := testWriteProjects(ctx, conn); err != nil {
		return 1
	}

	if err := testWriteCustomFields(ctx, conn); err != nil {
		return 1
	}

	if err := testWritePortfolios(ctx, conn); err != nil {
		return 1
	}

	if err := testWriteTasks(ctx, conn); err != nil {
		return 1
	}

	if err := testWriteTeams(ctx, conn); err != nil {
		return 1
	}

	return 0
}

func testWriteProjects(ctx context.Context, conn *cc.Connector) error {
	slog.Info("Creating projects...")

	params := common.WriteParams{
		ObjectName: "projects",
		RecordData: map[string]any{
			"name":         "Stuff to buy",
			"archived":     false,
			"color":        "light-green",
			"default_view": "calendar",
			"due_date":     "2019-09-15",
			"due_on":       "2019-09-15",
			"team":         "1209100536982881",
			"workspace":    "1206661566061885",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("error writing to Asana: %w", err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testWriteCustomFields(ctx context.Context, conn *cc.Connector) error {
	slog.Info("Testing write for custom_fields")

	params := common.WriteParams{
		ObjectName: "custom_fields",
		RecordData: map[string]any{
			"name":             gofakeit.BeerName(),
			"workspace":        "1206661566061885",
			"resource_subtype": "text",
		},
	}

	result, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to write custom_fields: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields result: %w", err)
	}

	os.Stdout.Write(data)
	os.Stdout.WriteString("\n")

	return nil
}

func testWritePortfolios(ctx context.Context, conn *cc.Connector) error {
	slog.Info("Testing write for portfolios")

	params := common.WriteParams{
		ObjectName: "portfolios",
		RecordData: map[string]any{
			"name":      "Test Portfolio",
			"workspace": "1206661566061885",
		},
	}

	result, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to write portfolios: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal portfolios result: %w", err)
	}

	os.Stdout.Write(data)
	os.Stdout.WriteString("\n")

	return nil
}

func testWriteTasks(ctx context.Context, conn *cc.Connector) error {
	slog.Info("Testing write for tasks")

	params := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"name":      "Test Task",
			"workspace": "1206661566061885",
		},
	}

	result, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to write tasks: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks result: %w", err)
	}

	os.Stdout.Write(data)
	os.Stdout.WriteString("\n")

	slog.Info("Testing update for tasks")

	updateParams := common.WriteParams{
		ObjectName: "tasks",
		RecordId:   result.RecordId,
		RecordData: map[string]any{
			"name": "Updated Test Task",
		},
	}

	updateResult, err := conn.Write(ctx, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update tasks: %w", err)
	}

	updateData, err := json.MarshalIndent(updateResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated tasks result: %w", err)
	}

	os.Stdout.Write(updateData)
	os.Stdout.WriteString("\n")

	return nil
}

func testWriteTeams(ctx context.Context, conn *cc.Connector) error {
	slog.Info("Testing write for teams")

	params := common.WriteParams{
		ObjectName: "teams",
		RecordData: map[string]any{
			"name":         gofakeit.Company(),
			"organization": "1206661566061885",
		},
	}

	result, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to write teams: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal teams result: %w", err)
	}

	os.Stdout.Write(data)
	os.Stdout.WriteString("\n")

	return nil
}
