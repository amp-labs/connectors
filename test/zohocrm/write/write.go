package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zohocrm.GetZohoConnector(ctx)
	defer utils.Close(conn)

	if err := createDeals(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createLeads(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func createDeals(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "Deals",
		RecordData: []map[string]any{
			{
				"id":        "3652397000003852095",
				"Deal_Name": "v6 Update",
				"Stage":     "Closed Won",
				"Pipeline":  "Standard (Standard)",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createLeads(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "Leads",
		RecordData: []map[string]any{
			{
				"Lead_Source": "Employee Referral",
				"Company":     "ABC",
				"Last_Name":   "Daly",
				"First_Name":  "Paul",
				"Email":       "p.daly@zylker.com",
				"State":       "Texas",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
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

func updateContacts(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "64934900000005440112",
		RecordData: []map[string]any{
			{
				"First_Name": "Ryan",
				"Phone":      "+12343678",
				"Last_Name":  "Dahl",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
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
