package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/loxo"
	"github.com/amp-labs/connectors/test/loxo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testPersonEvents(ctx)
	if err != nil {
		return 1
	}

	err = testSourceTypes(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testPersonEvents(ctx context.Context) error {
	conn := loxo.GetLoxoConnector(ctx)

	slog.Info("Deleting the person events")

	deleteParams := common.DeleteParams{
		ObjectName: "person_events",
		RecordId:   "4985924",
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

func testSourceTypes(ctx context.Context) error {
	conn := loxo.GetLoxoConnector(ctx)

	slog.Info("Deleting the source types")

	deleteParams := common.DeleteParams{
		ObjectName: "source_types",
		RecordId:   "38506",
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
