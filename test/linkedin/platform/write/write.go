package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/linkedin"
	"github.com/amp-labs/connectors/test/linkedin"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testPosts(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testPosts(ctx context.Context) error {
	conn := linkedin.GetPlatformConnector(ctx)

	slog.Info("Creating the posts")

	writeParams := common.WriteParams{
		ObjectName: "posts",
		RecordData: map[string]any{
			"author":     "urn:li:organization:2414183",
			"commentary": "Day 14 work",
			"visibility": "PUBLIC",
			"distribution": map[string]any{
				"feedDistribution": "MAIN_FEED",
			},
			"lifecycleState":            "PUBLISHED",
			"isReshareDisabledByAuthor": false,
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

	slog.Info("updating the posts")

	updateParams := common.WriteParams{
		ObjectName: "posts",
		RecordData: map[string]any{
			"patch": map[string]any{
				"$set": map[string]any{
					"commentary":               "Update to the post",
					"contentCallToActionLabel": "LEARN_MORE",
				},
				"adContext": map[string]any{
					"$set": map[string]any{
						"dscName": "Updating name!",
					},
				},
			},
		},
		RecordId: writeRes.RecordId,
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
