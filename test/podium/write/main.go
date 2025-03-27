package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	pd "github.com/amp-labs/connectors/providers/podium"
	"github.com/amp-labs/connectors/test/podium"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := podium.GetConnector(ctx)

	err := testCreatingContact(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateContact(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testUpdateContact(ctx context.Context, conn *pd.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "+17659807654",
		RecordData: map[string]any{
			"name": "John Smith",
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

func testCreatingContact(ctx context.Context, conn *pd.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"email": "john.doe10@podium.com",
			"locations": []string{
				"0195a414-7477-7484-a7fe-58bd9aaa3174",
			},
			"name":        "John Doe",
			"phoneNumber": "+18884441234",
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
