package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ds "github.com/amp-labs/connectors/providers/docusign"
	"github.com/amp-labs/connectors/test/docusign"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Use GetDocusignDeveloperConnector to connect to a developer account
	conn := docusign.GetDocusignConnector(ctx)
	authdata, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(authdata)

	if err := testReadTemplates(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadBulkSendBatch(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadBulkSendLists(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadEnvelopes(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadFolders(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadSigningGroups(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadCustomTabs(ctx, conn); err != nil {
		slog.Info(err.Error())
	}

	if err := testReadUsers(ctx, conn); err != nil {
		slog.Info(err.Error())
	}
}

func testReadTemplates(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "templates",
		Fields:     connectors.Fields("templateId", "description"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadBulkSendBatch(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "bulk_send_batch",
		Fields:     connectors.Fields("batchName", "action", "batchId"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadBulkSendLists(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "bulk_send_lists",
		Fields:     connectors.Fields("bulkSendListId", "name"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadEnvelopes(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "envelopes",
		Fields:     connectors.Fields("documentsUri", "envelopeId"),
		Since:      time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
		PageSize:   2,
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadFolders(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "folders",
		Fields:     connectors.Fields("folderId", "name"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadSigningGroups(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "signing_groups",
		Fields:     connectors.Fields("createdBy", "signingGroupId"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadCustomTabs(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "tab_definitions",
		Fields:     connectors.Fields("createdByUserId", "createdByDisplayName"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func testReadUsers(ctx context.Context, conn *ds.Connector) error {
	params := common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("email", "userName"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	return printReadResult(res)
}

func printReadResult(res *common.ReadResult) error {
	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")
	return nil
}
