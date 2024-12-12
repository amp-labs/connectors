package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/kit"
	"github.com/amp-labs/connectors/test/kit"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testCustomfields(ctx)
	if err != nil {
		return 1
	}

	err = testSubscribers(ctx)
	if err != nil {
		return 1
	}

	err = testTags(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testCustomfields(ctx context.Context) error {
	conn := kit.GetKitConnector(ctx)

	slog.Info("Creating the customfields")

	params := common.WriteParams{
		ObjectName: "custom_fields",
		RecordData: map[string]interface{}{
			"label": "customtest33",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, params)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the custom fields")

	updateParams := common.WriteParams{
		ObjectName: "custom_fields",
		RecordData: map[string]any{
			"label": "customtest33",
		},
		RecordId: writeRes.RecordId,
	}

	writeResp, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeResp); err != nil {
		return err
	}

	return nil
}

func testSubscribers(ctx context.Context) error {
	conn := kit.GetKitConnector(ctx)

	slog.Info("Creating the subscribers")

	params := common.WriteParams{
		ObjectName: "subscribers",
		RecordData: map[string]interface{}{
			"email_address": "dinesh@gmail.com",
			"fields": map[string]string{
				"Last name": "kumar",
				"Birthday":  "Mar 19",
				"Source":    "Landing page",
			},
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, params)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the subscribers")

	updateParams := common.WriteParams{
		ObjectName: "subscribers",
		RecordData: map[string]any{
			"first_name": "Dinesh",
			"fields": map[string]string{
				"Last name": "k",
			},
		},
		RecordId: writeRes.RecordId,
	}

	writeResp, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeResp); err != nil {
		return err
	}

	return nil
}

func testTags(ctx context.Context) error {
	conn := kit.GetKitConnector(ctx)

	slog.Info("Creating the tags")

	params := common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]interface{}{
			"name": "customer44",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, params)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the tags")

	updateParams := common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"name": "customer44",
		},
		RecordId: writeRes.RecordId,
	}

	writeResp, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeResp); err != nil {
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
