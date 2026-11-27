package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// Values of a person and its organization that exist in the test account.
// The phone must be given exactly as it is stored in Pipedrive.
const (
	personID       = "1"
	personEmail    = "aaronbidderman@gmail.com"
	personPhone    = "77051838136"
	organizationID = "1"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPipedriveConnector(ctx, providers.ModulePipedriveCRM)

	slog.Info("Searching person by id...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "persons",
		Fields:     connectors.Fields("id", "name", "emails", "phones", "org_id"),
		Filter:     equals("id", personID),
	})

	slog.Info("Searching person by email...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "persons",
		Fields:     connectors.Fields("id", "name", "emails", "phones", "org_id"),
		Filter:     equals("email", personEmail),
	})

	slog.Info("Searching person by phone...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "persons",
		Fields:     connectors.Fields("id", "name", "emails", "phones", "org_id"),
		Filter:     equals("phone", personPhone),
	})

	slog.Info("Searching organization by id...")
	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "organizations",
		Fields:     connectors.Fields("id", "name", "owner_id"),
		Filter:     equals("id", organizationID),
	})
}

func equals(fieldName string, value any) connectors.SearchFilter {
	return connectors.SearchFilter{
		FieldFilters: []common.FieldFilter{
			{
				FieldName: fieldName,
				Operator:  common.FilterOperatorEQ,
				Value:     value,
			},
		},
	}
}
