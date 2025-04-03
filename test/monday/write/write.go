package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/monday"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testWriteBoards(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testWriteBoards(ctx context.Context) error {
	conn := monday.GetMondayConnector(ctx)

	params := common.WriteParams{
		ObjectName: "boards",
		RecordData: map[string]any{
			"name":         gofakeit.Name(),
			"board_kind":   "public",
			"workspace_id": 1234567,
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		fmt.Println("ERR: ", err)
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
