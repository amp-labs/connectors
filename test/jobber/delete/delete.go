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

	err := testDeleteExpense(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testDeleteExpense(ctx context.Context) error {
	conn := jobber.GetJobberConnector(ctx)

	slog.Info("Delete Expense")

	deleteParams := common.DeleteParams{
		ObjectName: "expenses",
		RecordId:   "Z2lkOi8vSm9iYmVyL0V4cGVuc2UvMTU2NDQ3NTA=",
	}

	deleteRes, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(deleteRes); err != nil {
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
