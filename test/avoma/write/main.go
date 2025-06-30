package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/avoma"
	"github.com/amp-labs/connectors/test/avoma"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testSmartCategories(ctx)
	if err != nil {
		return 1
	}

	err = testCalls(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testSmartCategories(ctx context.Context) error {
	conn := avoma.GetAvomaConnector(ctx)

	slog.Info("Creating the smart categories")

	writeParams := common.WriteParams{
		ObjectName: "smart_categories",
		RecordData: map[string]any{
			"keywords": []string{
				"play",
			},
			"name": "sports",
			"prompts": []string{
				"football",
			},
			"settings": map[string]any{
				"aug_notes_enabled":       true,
				"keyword_notes_enabled":   true,
				"prompt_extract_length":   "short",
				"prompt_extract_strategy": "after",
			},
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

	slog.Info("update the smart categories")

	updateParams := common.WriteParams{
		ObjectName: "smart_categories",
		RecordData: map[string]any{
			"name": "Match",
			"prompts": []string{
				"circket",
			},
		},
		RecordId: "5d2830dd-7414-4bc2-81fa-6eec79917928",
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
		return err
	}
	return nil
}

func testCalls(ctx context.Context) error {
	conn := avoma.GetAvomaConnector(ctx)

	slog.Info("creating calls")

	writeParams := common.WriteParams{
		ObjectName: "calls",
		RecordData: map[string]any{
			"direction":   "Inbound",
			"external_id": "85127038997",
			"frm":         "+11234567857",
			"participants": []any{
				map[string]any{
					"email": "demo1@example.com",
				},
			},
			"recording_url": "https://example3.com/recording.mp3",
			"source":        "ringcentral",
			"start_at":      "2025-06-17T19:00:00Z",
			"to":            "+12234567890",
			"user_email":    "integration.test@withampersand.com",
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
