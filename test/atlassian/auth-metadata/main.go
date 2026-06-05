package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/atlassian"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
)

// Test script goal:
//
//	Call GetPostAuthInfo and confirm that the "cloudId" value can be retrieved.
//	This should work identically for any Atlassian module (Jira, Confluence).
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	slog.Info("Jira, fetch cloud id")
	jiraConn := connTest.GetJiraConnector(ctx)
	jiraCloudID := fetchCloudID(ctx, jiraConn)

	slog.Info("Confluence, fetch cloud id")
	confConn := connTest.GetConfluenceConnector(ctx)
	confCloudID := fetchCloudID(ctx, confConn)

	if jiraCloudID != confCloudID {
		utils.Fail("cloud ids for Jira and Confluence differ")
	}
}

func fetchCloudID(ctx context.Context, conn *atlassian.Connector) string {
	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	cloudId := (*info.CatalogVars)["cloudId"]

	if len(cloudId) == 0 {
		utils.Fail("missing cloud id in post authentication metadata")
	}

	slog.Info("retrieved auth metadata", "cloud id", cloudId)

	return cloudId
}
