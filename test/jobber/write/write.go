package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/jobber"
	"github.com/amp-labs/connectors/test/jobber"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testClient(ctx)
	if err != nil {
		return 1
	}

	err = testExpense(ctx)
	if err != nil {
		return 1
	}

	err = testProductAndServices(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testClient(ctx context.Context) error {
	conn := jobber.GetJobberConnector(ctx)

	slog.Info("Creating the client")

	writeParams := common.WriteParams{
		ObjectName: "clients",
		RecordData: map[string]any{
			"title":       "MR",
			"firstName":   "Deepak",
			"lastName":    "kumar",
			"companyName": "google",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the client")

	updateParams := common.WriteParams{
		ObjectName: "clients",
		RecordData: map[string]any{
			"isCompany": true,
			"emailsToAdd": map[string]any{
				"description": "MAIN",
				"address":     "deepakkumar@gmail.com",
				"primary":     true,
			},
		},
		RecordId: writeRes.RecordId,
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
		return err
	}

	return nil
}

func testExpense(ctx context.Context) error {
	conn := jobber.GetJobberConnector(ctx)

	slog.Info("Creating the expense")

	writeParams := common.WriteParams{
		ObjectName: "expenses",
		RecordData: map[string]any{
			"title":       "Today expense",
			"date":        "2025-09-12T10:45:30Z",
			"description": "expenses for today",
			"total":       100.00,
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the expense")

	updateParams := common.WriteParams{
		ObjectName: "expenses",
		RecordData: map[string]any{
			"total": 1000.00,
		},
		RecordId: writeRes.RecordId,
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
		return err
	}

	return nil
}

func testProductAndServices(ctx context.Context) error {
	conn := jobber.GetJobberConnector(ctx)

	slog.Info("Creating the products and services")

	writeParams := common.WriteParams{
		ObjectName: "productsAndServices",
		RecordData: map[string]any{
			"name":            "Mobile",
			"defaultUnitCost": 50000.00,
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the products and services")

	updateParams := common.WriteParams{
		ObjectName: "productsAndServices",
		RecordData: map[string]any{
			"defaultUnitCost": 10000.00,
		},
		RecordId: writeRes.RecordId,
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
		return err
	}

	return nil
}

func Write(ctx context.Context, conn *ap.Connector, payload common.WriteParams) (*common.WriteResult, error) {
	res, err := conn.Write(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// unmarshal the write response.
func constructResponse(res *common.WriteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
