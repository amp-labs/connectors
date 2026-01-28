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

	err := testLead(ctx)
	if err != nil {
		return 1
	}

	err = testReminders(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testLead(ctx context.Context) error {
	conn := breakcold.GetBreakcoldConnector(ctx)

	slog.Info("Deleting the lead")

	deleteParams := common.DeleteParams{
		ObjectName: "lead",
		RecordId:   "fe3b2787-dc05-4f45-942d-a45ee960d9",
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

func testReminders(ctx context.Context) error {
	conn := breakcold.GetBreakcoldConnector(ctx)

	slog.Info("Deleting the reminders")

	deleteParams := common.DeleteParams{
		ObjectName: "reminders",
		RecordId:   "be392c6c-d785-435b-9eb8-d4988f025160",
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
