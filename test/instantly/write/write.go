package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/test/instantly"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testLeadLabel(ctx)
	if err != nil {
		return 1
	}

	err = testCustomTags(ctx)
	if err != nil {
		return 1
	}

	err = testLeadLists(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testLeadLabel(ctx context.Context) error {
	conn := instantly.GetInstantlyConnector(ctx)

	slog.Info("Creating the lead labels")

	writeParams := common.WriteParams{
		ObjectName: "lead-labels",
		RecordData: map[string]any{
			"label":                 "Hot Lead",
			"interest_status_label": "positive",
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

	slog.Info("Updating the lead labels")

	updateParams := common.WriteParams{
		ObjectName: "lead-labels",
		RecordData: map[string]any{
			"label": "Lead",
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

func testCustomTags(ctx context.Context) error {
	conn := instantly.GetInstantlyConnector(ctx)

	slog.Info("Creating the custom tags")

	writeParams := common.WriteParams{
		ObjectName: "custom-tags",
		RecordData: map[string]any{
			"label": "Important",
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

	slog.Info("Updating the custom tags")

	updateParams := common.WriteParams{
		ObjectName: "custom-tags",
		RecordData: map[string]any{
			"label": "Demo",
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

func testLeadLists(ctx context.Context) error {
	conn := instantly.GetInstantlyConnector(ctx)

	slog.Info("Creating the lead lists")

	writeParams := common.WriteParams{
		ObjectName: "lead-lists",
		RecordData: map[string]any{
			"name": "Lead List",
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

	slog.Info("Updating the lead lists")

	updateParams := common.WriteParams{
		ObjectName: "lead-lists",
		RecordData: map[string]any{
			"name": "Demo",
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
