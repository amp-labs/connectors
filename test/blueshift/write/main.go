package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/blueshift"
	hs "github.com/amp-labs/connectors/test/blueshift"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetBlueshiftConnector(ctx)

	slog.Info("> TEST Create Customer")

	if err := createCustomer(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Create Custom User List")
	if err := createCustomUserList(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

}

func createCustomer(ctx context.Context, conn *blueshift.Connector) error {
	config := common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"email":       "test@gmail.com",
			"customer_id": "CUSTOMER_ID",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createCustomUserList(ctx context.Context, conn *blueshift.Connector) error {
	config := common.WriteParams{
		ObjectName: "custom_user_lists/create",
		RecordData: map[string]any{
			"name":         "need again new unique",
			"description":  "DESCRIPTION",
			"is_seed_list": "1",
			"source":       "SOURCE",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
