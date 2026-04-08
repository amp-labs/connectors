package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	testsupersend "github.com/amp-labs/connectors/test/supersend"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type labelPayload struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	TeamId string `json:"TeamId"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testsupersend.GetSuperSendConnector(ctx)

	// Create a team first (required for label creation)
	teamID := createTeam(ctx, conn)

	labelName := fmt.Sprintf("Delete Test Label %d", os.Getpid())

	testscenario.ValidateCreateDelete(ctx, conn,
		"labels",
		labelPayload{
			Name:   labelName,
			Color:  "#FF0000",
			TeamId: teamID,
		},
		testscenario.CRDTestSuite{
			ReadFields: datautils.NewSet("id", "name", "color"),
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: labelName,
			},
			RecordIdentifierKey: "id",
		},
	)
}

func createTeam(ctx context.Context, conn testscenario.ConnectorCRUD) string {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "teams",
		RecordData: map[string]any{
			"name":   fmt.Sprintf("Delete Test Team %d", os.Getpid()),
			"domain": "deleteteam.example.com",
			"about":  "Created for delete test",
		},
	})
	if err != nil {
		slog.Error("create team failed", "error", err)
		os.Exit(1)
	}

	return res.RecordId
}
