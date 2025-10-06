package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/fireflies"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testDeleteTranscripts(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testDeleteTranscripts(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Delete Transcripts")

	deleteParams := common.DeleteParams{
		ObjectName: "transcripts",
		RecordId:   "01JSXHYDZETZ24B9VEXQM52R9X",
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
