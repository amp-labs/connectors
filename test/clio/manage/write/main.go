package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	clioConn "github.com/amp-labs/connectors/providers/clio"
	connTest "github.com/amp-labs/connectors/test/clio"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

type expenseCategoryPayload struct {
	Name      string  `json:"name"`
	Rate      float64 `json:"rate"`
	EntryType string  `json:"entry_type"`
	UtbmsCode any     `json:"utbms_code"`
}

type groupPayload struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetClioManageConnector(ctx)

	slog.Info("Running expense_categories create/update/delete")
	runExpenseCategoriesCreateUpdateDelete(ctx, conn)

	slog.Info("Running groups create/update/delete")
	runGroupsCreateUpdateDelete(ctx, conn)
}

func runExpenseCategoriesCreateUpdateDelete(ctx context.Context, conn *clioConn.Connector) {
	name := "Scout Expense Category " + gofakeit.LetterN(8)
	updatedName := "Scout Expense Category Updated " + gofakeit.LetterN(8)

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "expense_categories",
		RecordData: expenseCategoryPayload{
			Name:      name,
			Rate:      0.1,
			EntryType: "hard_cost",
			UtbmsCode: nil,
		},
	})
	if err != nil {
		utils.Fail("error creating expense_category", "error", err)
	}

	if !createRes.Success {
		utils.Fail("failed to create expense_category", "response", createRes)
	}

	utils.DumpJSON(createRes, os.Stdout)

	recordID := createRes.RecordId
	if recordID == "" {
		utils.Fail("failed to create expense_category", "reason", "missing record id")
	}

	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "expense_categories",
		RecordId:   recordID,
		RecordData: expenseCategoryPayload{
			Name:      updatedName,
			Rate:      0.1,
			EntryType: "hard_cost",
			UtbmsCode: nil,
		},
	})
	if err != nil {
		utils.Fail("error updating expense_category", "error", err)
	}

	if !updateRes.Success {
		utils.Fail("failed to update expense_category", "response", updateRes)
	}

	utils.DumpJSON(updateRes, os.Stdout)

	delRes, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "expense_categories",
		RecordId:   recordID,
	})
	if err != nil {
		utils.Fail("error deleting expense_category", "error", err)
	}

	if !delRes.Success {
		utils.Fail("failed to delete expense_category", "response", delRes)
	}

	utils.DumpJSON(delRes, os.Stdout)
}

func runGroupsCreateUpdateDelete(ctx context.Context, conn *clioConn.Connector) {

	name := "Scout Group"
	updatedName := "Scout Group Updated"

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "groups",
		RecordData: groupPayload{
			Name: name,
			Type: "AdhocGroup",
		},
	})
	if err != nil {
		utils.Fail("error creating group", "error", err)
	}

	if !createRes.Success {
		utils.Fail("failed to create group", "response", createRes)
	}

	utils.DumpJSON(createRes, os.Stdout)

	recordID := createRes.RecordId
	if recordID == "" {
		utils.Fail("failed to create group", "reason", "missing record id")
	}

	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "groups",
		RecordId:   recordID,
		RecordData: groupPayload{
			Name: updatedName,
			Type: "AdhocGroup",
		},
	})
	if err != nil {
		utils.Fail("error updating group", "error", err)
	}

	if !updateRes.Success {
		utils.Fail("failed to update group", "response", updateRes)
	}

	utils.DumpJSON(updateRes, os.Stdout)

	delRes, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "groups",
		RecordId:   recordID,
	})
	if err != nil {
		utils.Fail("error deleting group", "error", err)
	}

	if !delRes.Success {
		utils.Fail("failed to delete group", "response", delRes)
	}

	utils.DumpJSON(delRes, os.Stdout)
}
