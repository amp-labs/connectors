package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	g2Conn "github.com/amp-labs/connectors/providers/g2"
	"github.com/amp-labs/connectors/test/g2"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := g2.NewConnector(ctx)

	err := testCreatingNewProductMapping(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingReview(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingNewProductMapping(ctx context.Context, conn *g2Conn.Connector) error {
	params := common.WriteParams{
		ObjectName: "product_mappings",
		RecordData: map[string]any{
			"external_id":          "string",
			"product_id":           "string",
			"match_score":          0,
			"partner_product_name": "string",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
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

func testCreatingReview(ctx context.Context, conn *g2Conn.Connector) error {
	params := common.WriteParams{
		ObjectName: "review",
		RecordData: map[string]any{
			"terms_accepted":          true,
			"privacy_policy_accepted": true,
			"answers": []map[string]any{
				{
					"question_id": "string",
					"value":       "string",
				},
			},
			"landing_page_id": "string",
			"source_type":     "string",
			"utm_param_id":    "string",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
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
