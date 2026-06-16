package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/odoo"
	connTest "github.com/amp-labs/connectors/test/odoo"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

// Odoo model used for the sequence; pick something your DB allows creating.
const testObjectName = "crm.iap.lead.role"

func main() {
	utils.SetupLogging()

	ctx := context.Background()
	conn := connTest.GetConnector(ctx)

	name := "amp-test-" + gofakeit.UUID()
	updated := "amp-updated-" + gofakeit.UUID()

	id := testCreate(ctx, conn, name)
	testUpdate(ctx, conn, id, updated)
	testDelete(ctx, conn, id)

	slog.Info("write-delete sequence finished", "object", testObjectName, "id", id)
}

func testCreate(ctx context.Context, conn *odoo.Connector, name string) string {
	params := common.WriteParams{
		ObjectName: testObjectName,
		RecordData: map[string]any{
			"name":      name,
			"reveal_id": "sale",
		},
	}

	slog.Info("creating record", "object", testObjectName)

	res, err := conn.Write(ctx, params)
	if err != nil {
		utils.Fail("create request failed", "error", err)
	}

	if !res.Success || res.RecordId == "" {
		utils.Fail("create failed", "response", res)
	}

	utils.DumpJSON(res, os.Stdout)

	return res.RecordId
}

func testUpdate(ctx context.Context, conn *odoo.Connector, id, updatedName string) {
	params := common.WriteParams{
		ObjectName: testObjectName,
		RecordId:   id,
		RecordData: map[string]any{
			"name": updatedName,
		},
	}

	slog.Info("updating record", "object", testObjectName, "id", id)

	res, err := conn.Write(ctx, params)
	if err != nil {
		utils.Fail("update request failed", "error", err)
	}

	if !res.Success {
		utils.Fail("update failed", "response", res)
	}

	utils.DumpJSON(res, os.Stdout)
}

func testDelete(ctx context.Context, conn *odoo.Connector, id string) {
	params := common.DeleteParams{
		ObjectName: testObjectName,
		RecordId:   id,
	}

	slog.Info("deleting record", "object", testObjectName, "id", id)

	res, err := conn.Delete(ctx, params)
	if err != nil {
		utils.Fail("delete request failed", "error", err)
	}

	if !res.Success {
		utils.Fail("delete failed", "response", res)
	}

	utils.DumpJSON(res, os.Stdout)
}
