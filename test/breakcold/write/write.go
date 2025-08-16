package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/breakcold"
	"github.com/amp-labs/connectors/test/breakcold"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testNotes(ctx)
	if err != nil {
		return 1
	}

	err = testLead(ctx)
	if err != nil {
		return 1
	}

	err = testReminders(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testNotes(ctx context.Context) error {
	conn := breakcold.GetBreakcoldConnector(ctx)

	slog.Info("Creating the notes")

	writeParams := common.WriteParams{
		ObjectName: "notes",
		RecordData: map[string]any{
			"content": "Enhancement meeting 14 Aug",
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

	slog.Info("Updating the notes")

	updateParams := common.WriteParams{
		ObjectName: "notes",
		RecordData: map[string]any{
			"content": "Enhancement meeting 14 Aug 2025",
		},
		RecordId: writeRes.RecordId,
	}

	res, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func testLead(ctx context.Context) error {
	conn := breakcold.GetBreakcoldConnector(ctx)

	slog.Info("Creating the lead")

	writeParams := common.WriteParams{
		ObjectName: "lead",
		RecordData: map[string]any{
			"company":      "lever",
			"company_role": "service provider",
			"first_name":   "Lever",
			"city":         "Califonia",
			"last_name":    "provider",
			"email":        "lever@gmail.com",
			"is_company":   true,
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

	slog.Info("Updating the lead")

	updateParams := common.WriteParams{
		ObjectName: "lead",
		RecordData: map[string]any{
			"data": map[string]any{
				"last_name": "lever",
			},
		},
		RecordId: writeRes.RecordId,
	}

	res, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func testReminders(ctx context.Context) error {
	conn := breakcold.GetBreakcoldConnector(ctx)

	slog.Info("Creating the reminders")

	writeParams := common.WriteParams{
		ObjectName: "reminders",
		RecordData: map[string]any{
			"name": "demo",
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

	slog.Info("Updating the reminders")

	updateParams := common.WriteParams{
		ObjectName: "reminders",
		RecordData: map[string]any{
			"name": "Meeting reminder",
		},
		RecordId: writeRes.RecordId,
	}

	res, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
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
