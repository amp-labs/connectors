package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	zi "github.com/amp-labs/connectors/providers/zoominfo"
	connTest "github.com/amp-labs/connectors/test/zoominfo"
)

// This harness exercises Write (create + update) across every writable ZoomInfo
// object, covering both write styles:
//   - GTM Studio (audiences, audience-folders): POST to create, PATCH /{id} to update.
//   - GTM Copilot config (customer-buyer-personas, customer-competitors,
//     ideal-company-profile, products): upsert via POST (id in body for update).
//
// NOTE: all write objects are entitlement-gated; without the relevant product the
// API returns 403. The harness logs each result/error and continues so every
// object is attempted (a 403, not a 404, still confirms request routing).
//
// Run: go run ./test/zoominfo/write
func main() {
	ctx := context.Background()
	conn := connTest.GetZoomInfoConnector(ctx)

	for _, test := range []func(context.Context, *zi.Connector) error{
		testCreateAudience,
		testUpdateAudience,
		testCreateAudienceFolder,
		testUpsertCustomerBuyerPersona,
		testUpsertCustomerCompetitor,
		testUpsertIdealCompanyProfile,
		testUpsertProduct,
	} {
		if err := test(ctx, conn); err != nil {
			slog.Error(err.Error())
		}
	}
}

func testCreateAudience(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "audiences",
		RecordData: map[string]any{"name": "Ampersand smoke test", "type": "CONTACT", "origin": "CUSTOM"},
	})
}

func testUpdateAudience(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "audiences",
		RecordId:   "550e8400-e29b-41d4-a716-446655440000",
		RecordData: map[string]any{"name": "Ampersand smoke test (renamed)"},
	})
}

func testCreateAudienceFolder(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "audience-folders",
		RecordData: map[string]any{"name": "Ampersand smoke folder", "description": "created by smoke test"},
	})
}

func testUpsertCustomerBuyerPersona(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "customer-buyer-personas",
		RecordData: map[string]any{"name": "Smoke Persona", "description": "created by smoke test"},
	})
}

func testUpsertCustomerCompetitor(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "customer-competitors",
		RecordData: map[string]any{"name": "Smoke Competitor"},
	})
}

func testUpsertIdealCompanyProfile(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "ideal-company-profile",
		RecordData: map[string]any{"name": "Smoke ICP"},
	})
}

func testUpsertProduct(ctx context.Context, conn *zi.Connector) error {
	return write(ctx, conn, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{"name": "Smoke Product"},
	})
}

func write(ctx context.Context, conn *zi.Connector, params common.WriteParams) error {
	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", params.ObjectName, err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
