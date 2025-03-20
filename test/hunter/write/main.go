package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	hr "github.com/amp-labs/connectors/providers/hunter"
	"github.com/amp-labs/connectors/test/hunter"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := hunter.GetHunterConnector(ctx)

	err := testCreatingLead(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateLead(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingLead(ctx context.Context, conn *hr.Connector) error {
	params := common.WriteParams{
		ObjectName: "leads",
		RecordData: map[string]any{
			"email":            gofakeit.Email(),
			"first_name":       "Aloyce",
			"last_name":        "Ohanian",
			"position":         "Cofounder",
			"company":          "RedEye",
			"company_industry": "Internet & Telecom",
			"company_size":     "20-30 employees",
			"confidence_score": 97,
			"website":          "reddit.com",
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

func testUpdateLead(ctx context.Context, conn *hr.Connector) error {
	params := common.WriteParams{
		ObjectName: "leads",
		RecordId:   "187693047",
		RecordData: map[string]any{
			"company": "Facebook",
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
