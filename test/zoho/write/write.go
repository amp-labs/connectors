package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoCRM)

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

func createDeals(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "Deals",
		RecordData: map[string]any{
			"id":        "3652397000003852095",
			"deal_name": "v6 Update",
			"stage":     "Closed Won",
			"pipeline":  "Standard (Standard)",
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

func createLeads(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "leads",
		RecordData: []map[string]any{
			{
				"lead_source": "Employee Referral",
				"company":     "ABC",
				"last_name":   "Daly",
				"first_name":  "Paul",
				"email":       "p.daly@zylker.com",
				"state":       "Texas",
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

func updateContacts(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "6172731000000472189",
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
