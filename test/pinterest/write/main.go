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

	err = testMedia(ctx)
	if err != nil {
		return 1
	}

	err = testCatalogs(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testBoards(ctx context.Context) error {
	conn := pinterest.GetConnector(ctx)

	slog.Info("Creating the boards")

	writeParams := common.WriteParams{
		ObjectName: "boards",
		RecordData: map[string]any{
			"name":        "Collection",
			"description": "collectoion of flowers",
			"privacy":     "PUBLIC",
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

	slog.Info("Updating the boards")

	updateParams := common.WriteParams{
		ObjectName: "boards",
		RecordData: map[string]any{
			"data": map[string]interface{}{
				"name": "Collections",
			},
		},
		RecordId: writeRes.Data["id"].(string),
	}

	writeres, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeres); err != nil {
		return err
	}

	return nil
}

func testMedia(ctx context.Context) error {
	conn := pinterest.GetConnector(ctx)

	slog.Info("Creating the media")

	writeParams := common.WriteParams{
		ObjectName: "media",
		RecordData: map[string]any{
			"media_type": "video",
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

func testCatalogs(ctx context.Context) error {
	conn := pinterest.GetConnector(ctx)

	slog.Info("Creating the catalogs")

	writeParams := common.WriteParams{
		ObjectName: "catalogs",
		RecordData: map[string]any{
			"catalog_type": "HOTEL",
			"name":         "DishWorld",
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
