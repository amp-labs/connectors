package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/pinterest"
	"github.com/amp-labs/connectors/test/pinterest"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testBoards(ctx)
	if err != nil {
		return 1
	}

	err = testPins(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testBoards(ctx context.Context) error {
	conn := pinterest.GetConnector(ctx)

	slog.Info("Deleting the boards")

	deleteParams := common.DeleteParams{
		ObjectName: "boards",
		RecordId:   "1048283319454998737",
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

func testPins(ctx context.Context) error {
	conn := pinterest.GetConnector(ctx)

	slog.Info("Deleting the pins")

	deleteParams := common.DeleteParams{
		ObjectName: "pins",
		RecordId:   "1048283250767720066",
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
