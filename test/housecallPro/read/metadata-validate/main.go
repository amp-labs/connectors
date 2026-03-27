package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	for _, objectName := range []string{
		"customers",
		"employees",
		"estimates",
		"jobs",
		"job_fields/job_types",
		"leads",
		"lead_sources",
		"price_book/material_categories",
		"price_book/price_forms",
		"price_book/services",
		"service_zones",
		"events",
		"tags",
		"invoices",
	} {
		testscenario.ValidateMetadataContainsRead(ctx, conn, objectName, nil)
	}
}
