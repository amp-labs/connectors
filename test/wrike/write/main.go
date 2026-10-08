package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/wrike"
)

// Exercises an account-level create and update (job roles) and an update of
// an existing task, reverted afterwards. Delete is not part of this step, so
// the created job role is left behind; remove it with DELETE /v4/jobroles/{id}.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetWrikeConnector(ctx)

	slog.Info("> TEST Create job role")

	role, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "jobroles",
		RecordData: map[string]any{"title": "Connector Test Role", "shortTitle": "CT"},
	})
	if err != nil {
		utils.Fail("error creating job role", "error", err)
	}

	utils.DumpJSON(role, os.Stdout)

	slog.Info("> TEST Update job role", "id", role.RecordId)

	updatedRole, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "jobroles",
		RecordId:   role.RecordId,
		RecordData: map[string]any{"title": "Connector Test Role (updated)"},
	})
	if err != nil {
		utils.Fail("error updating job role", "error", err)
	}

	utils.DumpJSON(updatedRole, os.Stdout)

	slog.Info("> TEST Update an existing task and revert")

	tasks, err := conn.Read(ctx, common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields("id", "title"), PageSize: 1})
	if err != nil || len(tasks.Data) == 0 {
		utils.Fail("error reading tasks", "error", err)
	}

	task := tasks.Data[0]
	originalTitle, _ := task.Fields["title"].(string)

	renamed, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tasks",
		RecordId:   task.Id,
		RecordData: map[string]any{"title": originalTitle + " (connector test)"},
	})
	if err != nil {
		utils.Fail("error updating task", "error", err)
	}

	utils.DumpJSON(renamed, os.Stdout)

	if _, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "tasks",
		RecordId:   task.Id,
		RecordData: map[string]any{"title": originalTitle},
	}); err != nil {
		utils.Fail("error reverting task title", "error", err)
	}

	slog.Info("task title reverted", "title", originalTitle, "jobRoleId", role.RecordId)
}
