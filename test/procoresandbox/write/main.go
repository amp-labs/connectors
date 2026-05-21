package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/procore"
	"github.com/amp-labs/connectors/test/procoresandbox"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn, err := procoresandbox.NewConnector(ctx)
	if err != nil {
		return fmt.Errorf("error creating connector: %w", err)
	}

	programId, err := testCreatingPrograms(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdatePrograms(ctx, conn, programId); err != nil {
		return err
	}

	roleId, err := testCreatingRoles(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateRoles(ctx, conn, roleId); err != nil {
		return err
	}

	return nil
}

func testCreatingPrograms(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "programs",
		RecordData: map[string]any{

			"program": map[string]any{
				"name":             gofakeit.Company(),
				"address_freeform": "500 Construction Way, Santa Barbara",
				"website":          "http://www.example.com",
				"zip":              91013,
			},
		},
	}

	slog.Info("Creating programs...")

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

func testUpdatePrograms(ctx context.Context, conn *cc.Connector, programId string) error {
	params := common.WriteParams{
		ObjectName: "programs",
		RecordId:   programId,
		RecordData: map[string]any{
			"name": "Updated " + gofakeit.Company(),
		},
	}

	slog.Info("Updating program...")

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

func testCreatingRoles(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "roles",
		RecordData: map[string]any{
			"project_role": map[string]any{
				"id":                      "12345",
				"add_to_project_team":     true,
				"archetype":               "owner",
				"display_on_company_home": true,
				"name":                    gofakeit.JobTitle(),
				"type":                    "contact",
			},
		},
	}

	slog.Info("Creating roles...")

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

func testUpdateRoles(ctx context.Context, conn *cc.Connector, roleId string) error {
	params := common.WriteParams{
		ObjectName: "roles",
		RecordId:   roleId,
		RecordData: map[string]any{
			"project_role": map[string]any{
				"name":      "Updated " + gofakeit.JobTitle(),
				"archetype": "owner",
				"type":      "contact",
			},
		},
	}

	slog.Info("Updating role...")

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
