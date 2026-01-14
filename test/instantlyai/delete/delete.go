package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/instantlyai"
	"github.com/amp-labs/connectors/test/instantlyai"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testCustomTags(ctx)
	if err != nil {
		return 1
	}

	err = testLeadLists(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testCustomTags(ctx context.Context) error {
	conn := instantlyai.GetInstantlyAIConnector(ctx)

	slog.Info("Deleting the custom tags")

	deleteParams := common.DeleteParams{
		ObjectName: "custom-tags",
		RecordId:   "4fa0ca9d-f205-4a68-8756-5ad62123a53a",
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

func testLeadLists(ctx context.Context) error {
	conn := instantlyai.GetInstantlyAIConnector(ctx)

	slog.Info("Deleting the lead lists")

	deleteParams := common.DeleteParams{
		ObjectName: "lead-lists",
		RecordId:   "3291573a-2eb9-4571-8e07-93d421774cc6",
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
