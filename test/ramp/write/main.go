package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/ramp"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRampConnector(ctx)

	slog.Info("> TEST Create department")

	createResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordData: map[string]any{
			"name": "Test Department (connector test)",
		},
	})
	if err != nil {
		utils.Fail("error creating department", "error", err)
	}

	utils.DumpJSON(createResult, os.Stdout)

	if createResult.RecordId == "" {
		utils.Fail("expected a record ID after create")
	}

	slog.Info("> TEST Update department", "id", createResult.RecordId)

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordId:   createResult.RecordId,
		RecordData: map[string]any{
			"name": "Test Department (updated)",
		},
	})
	if err != nil {
		utils.Fail("error updating department", "error", err)
	}

	utils.DumpJSON(updateResult, os.Stdout)

	slog.Info("> TEST Create location")

	locResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "locations",
		RecordData: map[string]any{
			"name": "Test Location (connector test)",
		},
	})
	if err != nil {
		utils.Fail("error creating location", "error", err)
	}

	utils.DumpJSON(locResult, os.Stdout)

	// Object names follow the Ramp API paths, so this one is hyphenated
	// ("spend-programs", not "spend_programs"). The OAuth scope keeps the
	// underscore form (spend_programs:write) -- the two are unrelated.
	slog.Info("> TEST Create spend program")

	programResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "spend-programs",
		RecordData: map[string]any{
			"display_name":                  "Test Spend Program (connector test)",
			"description":                   "Created by the ramp connector write test",
			"icon":                          "SoftwareTrialIcon",
			"is_shareable":                  false,
			"issue_physical_card_if_needed": false,
			"permitted_spend_types": map[string]any{
				"primary_card_enabled":   true,
				"reimbursements_enabled": false,
			},
			// Amounts are in the smallest currency unit, so this is $500.00.
			"spending_restrictions": map[string]any{
				"interval": "MONTHLY",
				"limit": map[string]any{
					"amount":        50000,
					"currency_code": "USD",
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error creating spend program", "error", err)
	}

	utils.DumpJSON(programResult, os.Stdout)

	if programResult.RecordId == "" {
		utils.Fail("expected a record ID after creating spend program")
	}

	slog.Info("Done")
}
