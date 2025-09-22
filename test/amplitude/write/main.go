package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/amplitude"
	"github.com/amp-labs/connectors/test/amplitude"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := amplitude.GetAmplitudeConnector(ctx)

	_, err := testCreatingAnnotations(ctx, conn)
	if err != nil {
		return err
	}

	_, err = testCreatingRelease(ctx, conn)
	if err != nil {
		return err
	}

	_, err = testCreatingAttribution(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAttribution(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "attribution",
		RecordData: map[string]any{
			"event_type": "[YOUR COMPANY] Install",
			"idfa":       "AEBE52E7-03EE-455A-B3C4-E57283966239",
			"user_properties": map[string]any{
				"[YOUR COMPANY] media source": "facebook",
				"[YOUR COMPANY] campaign":     "refer-a-friend",
			},
			"platform": "ios",
		},
	}

	slog.Info("Creating an attribution...")

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

	return res.RecordId, nil
}

func testCreatingAnnotations(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "annotations",
		RecordData: map[string]any{
			"app_id": "679680",
			"date":   "2025-09-16",
			"label":  "Version 2.4 Release",
		},
	}

	slog.Info("Creating an annotation...")

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

	return res.RecordId, nil
}

func testCreatingRelease(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "release",
		RecordData: map[string]any{
			"version":       gofakeit.AppVersion(),
			"release_start": "2025-12-01 00:00:00",
			"title":         "Version 2. Release",
		},
	}

	slog.Info("Creating a release...")

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

	return res.RecordId, nil
}
