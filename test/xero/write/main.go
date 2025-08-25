package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/xero"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/xero"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	// Set up slog logging.
	utils.SetupLogging()

	conn := xero.GetXeroConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)

	if err != nil {
		utils.Fail(err.Error())
	}

	recordId, err := testCreatingContactGroups(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateContactGroups(ctx, conn, recordId)
	if err != nil {
		return err
	}

	_, err = testCreatingTaxRates(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingContactGroups(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "contactGroups",
		RecordData: map[string]any{
			"name": gofakeit.Name(),
		},
	}
	slog.Info("Creating an contact group...")

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

func testUpdateContactGroups(ctx context.Context, conn *cc.Connector, recordId string) error {
	params := common.WriteParams{
		ObjectName: "contactGroups",
		RecordId:   recordId,
		RecordData: map[string]any{
			"ContactGroupID": recordId,
			"name":           gofakeit.Name(),
		},
	}
	slog.Info("Updating a contact group...")

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

func testCreatingTaxRates(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "taxRates",
		RecordData: map[string]any{
			"Name": gofakeit.Name(),
			"TaxComponents": []map[string]any{
				{
					"Name":             "Society Tax",
					"Rate":             "7.6",
					"IsCompound":       "false",
					"IsNonRecoverable": "false",
				},
			},
		},
	}
	slog.Info("Creating an tax rate...")

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
