package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zohocrm"
	"github.com/amp-labs/connectors/test/utils"
	testConn "github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testConn.GetZohoConnector(ctx)

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

func createDeals(ctx context.Context, conn *zohocrm.Connector) error {
	config := common.WriteParams{
		ObjectName: "Deals",
		RecordData: map[string]any{
			"id":        "3652397000003852095",
			"Deal_Name": "v6 Update",
			"Stage":     "Closed Won",
			"Pipeline":  "Standard (Standard)",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createLeads(ctx context.Context, conn *zohocrm.Connector) error {
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

func updateContacts(ctx context.Context, conn *zohocrm.Connector) error {
	config := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "6493490000001291001",
		RecordData: map[string]any{
			"First_Name": "Ryan",
			"Phone":      "+12343678910",
			"Last_Name":  "Dahl2",
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
