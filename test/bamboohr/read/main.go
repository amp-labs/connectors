package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/bamboohr"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBambooHRConnector(ctx)

	tests := []common.ReadParams{
		{
			ObjectName: "employees_directory",
			Fields:     connectors.Fields("id", "firstName", "lastName", "jobTitle"),
		},
		{
			ObjectName: "meta_users",
			Fields:     connectors.Fields("id", "employeeId", "email", "status"),
		},
		{
			ObjectName: "employees",
			Fields:     connectors.Fields("employeeId", "firstName", "lastName", "status"),
			PageSize:   100,
		},
		{
			ObjectName: "whos_out",
			Fields:     connectors.Fields("id", "employeeId", "start", "end"),
			PageSize:   100,
		},
		{
			ObjectName: "applicant_tracking_jobs",
			Fields:     connectors.Fields("id", "title", "status"),
		},
	}

	for _, tt := range tests {
		slog.Info("=== Reading " + tt.ObjectName + " ===")
		testscenario.ReadThroughPages(ctx, conn, tt)
	}
}
