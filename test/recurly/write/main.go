package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/recurly"
	"github.com/amp-labs/connectors/test/recurly"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := recurly.GetRecurlyConnector(ctx)

	itemId, err := testCreatingItems(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateItems(ctx, conn, itemId); err != nil {
		return err
	}

	couponsId, err := testCreatingCoupons(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateCoupons(ctx, conn, couponsId); err != nil {
		return err
	}

	return nil
}

func testCreatingItems(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "items",
		RecordData: map[string]any{
			"code": gofakeit.UUID(),
			"name": gofakeit.ProductName(),
		},
	}

	slog.Info("Creating items...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["id"].(string), nil
}

func testUpdateItems(ctx context.Context, conn *cc.Connector, itemId string) error {
	params := common.WriteParams{
		ObjectName: "items",
		RecordId:   itemId,
		RecordData: map[string]any{
			"name": "Updated " + gofakeit.ProductName(),
		},
	}

	slog.Info("Updating item...")

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

func testCreatingCoupons(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "coupons",
		RecordData: map[string]any{
			"code":              gofakeit.UUID(),
			"name":              gofakeit.ProductName(),
			"discount_type":     "free_trial",
			"free_trial_unit":   "day",
			"free_trial_amount": 14,
		},
	}

	slog.Info("Creating coupons...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["id"].(string), nil
}

func testUpdateCoupons(ctx context.Context, conn *cc.Connector, couponId string) error {
	params := common.WriteParams{
		ObjectName: "coupons",
		RecordId:   couponId,
		RecordData: map[string]any{
			"name": "Updated " + gofakeit.ProductName(),
		},
	}

	slog.Info("Updating coupon...")

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
