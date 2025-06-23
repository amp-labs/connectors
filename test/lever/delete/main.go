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

	err := testNotes(ctx)
	if err != nil {
		return 1
	}

	err = testRequisitionFields(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testNotes(ctx context.Context) error {
	conn := lever.GetConnector(ctx)

	slog.Info("Deleting the notes")

	deleteParams := common.DeleteParams{
		ObjectName: "notes",
		RecordId:   "f4b0dfe1-966e-4cbe-b4c8-c9c864be98d6",
	}

	res, err := Delete(ctx, conn, deleteParams)
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

	slog.Info("Deleting the requistion_fields")

	deleteParams := common.DeleteParams{
		ObjectName: "requisition_fields",
		RecordId:   "field2",
	}

	res, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, conn *ap.Connector, payload common.DeleteParams) (*common.DeleteResult, error) {
	res, err := conn.Delete(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// unmarshal the delte response.
func constructResponse(res *common.DeleteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
