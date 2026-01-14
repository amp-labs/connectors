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

	return 0
}

func testRequisitions(ctx context.Context) error {
	conn := lever.GetConnector(ctx)

	slog.Info("Deleting the requisitions")

	deleteParams := common.DeleteParams{
		ObjectName: "requisitions",
		RecordId:   "8998678f-a76c-4a03-9a19-24ffd20802e8",
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
		RecordId:   "field1",
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

// unmarshal the delete response.
func constructResponse(res *common.DeleteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
