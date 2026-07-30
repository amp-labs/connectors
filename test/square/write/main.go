package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/square"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSquareConnector(ctx)

	// Create a customer, then update it with the returned record id.
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"given_name":    "Amelia",
			"family_name":   "Earhart",
			"email_address": "amelia.earhart@example.com",
			"note":          "created by the connectors write test",
		},
	})
	if err != nil {
		utils.Fail("error creating customer", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordId:   res.RecordId,
		RecordData: map[string]any{
			"note": "updated by the connectors write test",
		},
	})
	if err != nil {
		utils.Fail("error updating customer", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	// Create a customer group, then update it.
	res, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "customers/groups",
		RecordData: map[string]any{
			"name": "Connectors Write Test Group",
		},
	})
	if err != nil {
		utils.Fail("error creating customer group", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "customers/groups",
		RecordId:   res.RecordId,
		RecordData: map[string]any{
			"name": "Connectors Write Test Group (updated)",
		},
	})
	if err != nil {
		utils.Fail("error updating customer group", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	// Create a digital gift card; exercises envelope wrapping, the hoisted
	// location_id field, and idempotency key injection.
	locations, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "locations",
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading locations", "error", err)
	}

	if len(locations.Data) == 0 {
		utils.Fail("no locations available for gift card creation")
	}

	locationID, ok := locations.Data[0].Fields["id"].(string)
	if !ok {
		utils.Fail("location id is not a string")
	}

	res, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "gift_cards",
		RecordData: map[string]any{
			"type":        "DIGITAL",
			"location_id": locationID,
		},
	})
	if err != nil {
		utils.Fail("error creating gift card", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Write operation completed successfully.")
}
