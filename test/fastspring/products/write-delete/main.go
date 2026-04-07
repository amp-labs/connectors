package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetFastSpringConnector(ctx)

	// Product path must look like a slug (e.g. "premium-laptop"); dots / odd shapes often fail validation.
	// See: https://developer.fastspring.com/reference/create-or-update-products
	productPath := fmt.Sprintf("amp-integration-%s", strings.ReplaceAll(gofakeit.UUID(), "-", ""))

	createPayload := map[string]any{
		"product": productPath,
		"display": map[string]any{
			"en": fmt.Sprintf("Amp integration %s", gofakeit.Word()),
		},
		"format": "digital",
		"description": map[string]any{
			"summary": map[string]any{
				"en": "Temporary connector integration test product.",
			},
		},
		"pricing": map[string]any{
			"price": map[string]any{
				"USD": 1.00,
			},
		},
	}

	testscenario.ValidateCreateDelete(ctx, conn, "products", createPayload,
		testscenario.CRDTestSuite{
			ReadFields: datautils.NewSet("path"),
			SearchBy: testscenario.Property{
				Key:   "path",
				Value: productPath,
			},
			RecordIdentifierKey: "path",
		},
	)
}
