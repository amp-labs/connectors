package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/lever"
	"github.com/amp-labs/connectors/test/lever"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testRequisitions(ctx)
	if err != nil {
		return 1
	}

	err = testRequisitionFields(ctx)
	if err != nil {
		return 1
	}

	err = testNotes(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testRequisitions(ctx context.Context) error {
	conn := lever.GetConnector(ctx)

	slog.Info("Creating the Requisitions")

	writeParams := common.WriteParams{
		ObjectName: "requisitions",
		RecordData: map[string]any{
			"requisitionCode": "ENG-20",
			"name":            "software developer",
			"headcountTotal":  10,
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

	slog.Info("Updating the Requisitions")

	updateParams := common.WriteParams{
		ObjectName: "requisitions",
		RecordData: map[string]any{
			"requisitionCode": "ENG-20",
			"headcountTotal":  10,
			"status":          "closed",
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

func testRequisitionFields(ctx context.Context) error {
	conn := lever.GetConnector(ctx)

	slog.Info("Creating the requisitionFields")

	writeParams := common.WriteParams{
		ObjectName: "requisition_fields",
		RecordData: map[string]any{
			"id":         "offboard_date",
			"text":       "Offboard date",
			"type":       "date",
			"isRequired": true,
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

	slog.Info("Updating the requisition_fields")

	updateParams := common.WriteParams{
		ObjectName: "requisition_fields",
		RecordData: map[string]any{
			"id":         "offboard_date",
			"text":       "Offboard date",
			"type":       "date",
			"isRequired": false,
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

func testNotes(ctx context.Context) error {
	conn := lever.GetConnector(ctx)

	slog.Info("Creating the notes")

	writeParams := common.WriteParams{
		ObjectName: "notes",
		RecordData: map[string]any{
			"value": "Hiring on 2+ experience",
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
